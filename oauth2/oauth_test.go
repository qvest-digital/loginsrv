package oauth2

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var testConfig = Config{
	ClientID:     "client42",
	ClientSecret: "secret",
	AuthURL:      "http://auth-provider/auth",
	TokenURL:     "http://auth-provider/token",
	RedirectURI:  "http://localhost/callback",
	Scope:        "email other",
}

func Test_StartFlow(t *testing.T) {
	resp := httptest.NewRecorder()
	StartFlow(testConfig, resp)

	assert.Equal(t, http.StatusFound, resp.Code)

	// assert that we received a state cookie
	cHeader := strings.Split(resp.Header().Get("Set-Cookie"), ";")[0]
	assert.Equal(t, stateCookieName, strings.Split(cHeader, "=")[0])
	state := strings.Split(cHeader, "=")[1]

	expectedLocation := fmt.Sprintf("%v?client_id=%v&redirect_uri=%v&response_type=code&scope=%v&state=%v",
		testConfig.AuthURL,
		testConfig.ClientID,
		url.QueryEscape(testConfig.RedirectURI),
		"email+other",
		state,
	)

	assert.Equal(t, expectedLocation, resp.Header().Get("Location"))
}

func Test_Authenticate(t *testing.T) {
	// mock a server for token exchange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		body, _ := ioutil.ReadAll(r.Body)
		assert.Equal(t, "client_id=client42&client_secret=secret&code=theCode", string(body))

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"e72e16c7e42f292c6912e7710c838347ae178b4a", "scope":"repo gist", "token_type":"bearer"}`))
	}))
	defer server.Close()

	testConfigCopy := testConfig
	testConfigCopy.TokenURL = server.URL

	request, _ := http.NewRequest("GET", testConfig.RedirectURI, nil)
	request.Header.Set("Cookie", "oauthState=theState")
	request.URL, _ = url.Parse("http://localhost/callback?code=theCode&state=theState")

	tokenInfo, err := Authenticate(testConfigCopy, request)

	assert.NoError(t, err)
	assert.Equal(t, "e72e16c7e42f292c6912e7710c838347ae178b4a", tokenInfo.AccessToken)
	assert.Equal(t, "repo gist", tokenInfo.Scope)
	assert.Equal(t, "bearer", tokenInfo.TokenType)
}

func Test_Authentication_ProviderError(t *testing.T) {
	request, _ := http.NewRequest("GET", testConfig.RedirectURI, nil)
	request.URL, _ = url.Parse("http://localhost/callback?error=provider_login_error")

	_, err := Authenticate(testConfig, request)

	assert.Error(t, err)
	assert.Equal(t, "error: provider_login_error", err.Error())
}

func Test_Authentication_StateError(t *testing.T) {
	request, _ := http.NewRequest("GET", testConfig.RedirectURI, nil)
	request.Header.Set("Cookie", "oauthState=XXXXXXX")
	request.URL, _ = url.Parse("http://localhost/callback?code=theCode&state=theState")

	_, err := Authenticate(testConfig, request)

	assert.Error(t, err)
	assert.Equal(t, "error: oauth state param could not be verified", err.Error())
}

func Test_Authentication_NoCodeError(t *testing.T) {
	request, _ := http.NewRequest("GET", testConfig.RedirectURI, nil)
	request.Header.Set("Cookie", "oauthState=theState")
	request.URL, _ = url.Parse("http://localhost/callback?state=theState")

	_, err := Authenticate(testConfig, request)

	assert.Error(t, err)
	assert.Equal(t, "error: no auth code provided", err.Error())
}

func Test_Authentication_Provider500(t *testing.T) {
	// mock a server for token exchange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	testConfigCopy := testConfig
	testConfigCopy.TokenURL = server.URL

	request, _ := http.NewRequest("GET", testConfig.RedirectURI, nil)
	request.Header.Set("Cookie", "oauthState=theState")
	request.URL, _ = url.Parse("http://localhost/callback?code=theCode&state=theState")

	_, err := Authenticate(testConfigCopy, request)

	assert.Error(t, err)
	assert.Equal(t, "error: expected http status 200 on token exchange, but got 500", err.Error())
}

func Test_Authentication_ProviderNetworkError(t *testing.T) {

	testConfigCopy := testConfig
	testConfigCopy.TokenURL = "http://localhost:12345678"

	request, _ := http.NewRequest("GET", testConfig.RedirectURI, nil)
	request.Header.Set("Cookie", "oauthState=theState")
	request.URL, _ = url.Parse("http://localhost/callback?code=theCode&state=theState")

	_, err := Authenticate(testConfigCopy, request)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid port")
}

func Test_Authentication_TokenParseError(t *testing.T) {
	// mock a server for token exchange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_t`))

	}))
	defer server.Close()

	testConfigCopy := testConfig
	testConfigCopy.TokenURL = server.URL

	request, _ := http.NewRequest("GET", testConfig.RedirectURI, nil)
	request.Header.Set("Cookie", "oauthState=theState")
	request.URL, _ = url.Parse("http://localhost/callback?code=theCode&state=theState")

	_, err := Authenticate(testConfigCopy, request)

	assert.Error(t, err)
	assert.Equal(t, "error on parsing oauth token: unexpected EOF", err.Error())
}
