package oauth2

import (
	"fmt"
	. "github.com/stretchr/testify/assert"
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

	Equal(t, http.StatusFound, resp.Code)

	// assert that we received a state cookie
	cHeader := strings.Split(resp.Header().Get("Set-Cookie"), ";")[0]
	Equal(t, stateCookieName, strings.Split(cHeader, "=")[0])
	state := strings.Split(cHeader, "=")[1]

	expectedLocation := fmt.Sprintf("%v?client_id=%v&redirect_uri=%v&response_type=code&scope=%v&state=%v",
		testConfig.AuthURL,
		testConfig.ClientID,
		url.QueryEscape(testConfig.RedirectURI),
		"email+other",
		state,
	)

	Equal(t, expectedLocation, resp.Header().Get("Location"))
}

func Test_Authenticate(t *testing.T) {
	// mock a server for token exchange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Equal(t, "POST", r.Method)
		Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		Equal(t, "application/json", r.Header.Get("Accept"))

		body, _ := ioutil.ReadAll(r.Body)
		Equal(t, "client_id=client42&client_secret=secret&code=theCode", string(body))

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

	NoError(t, err)
	Equal(t, "e72e16c7e42f292c6912e7710c838347ae178b4a", tokenInfo.AccessToken)
	Equal(t, "repo gist", tokenInfo.Scope)
	Equal(t, "bearer", tokenInfo.TokenType)
}

func Test_Authenticate_CodeExchangeError(t *testing.T) {
	var testReturnCode int
	testResponseJson := `{"error":"bad_verification_code","error_description":"The code passed is incorrect or expired.","error_uri":"https://developer.github.com/v3/oauth/#bad-verification-code"}`
	// mock a server for token exchange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(testReturnCode)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(testResponseJson))
	}))
	defer server.Close()

	testConfigCopy := testConfig
	testConfigCopy.TokenURL = server.URL

	request, _ := http.NewRequest("GET", testConfig.RedirectURI, nil)
	request.Header.Set("Cookie", "oauthState=theState")
	request.URL, _ = url.Parse("http://localhost/callback?code=theCode&state=theState")

	testReturnCode = 500
	tokenInfo, err := Authenticate(testConfigCopy, request)
	Error(t, err)
	EqualError(t, err, "error: expected http status 200 on token exchange, but got 500")
	Equal(t, "", tokenInfo.AccessToken)

	testReturnCode = 200
	tokenInfo, err = Authenticate(testConfigCopy, request)
	Error(t, err)
	EqualError(t, err, `error: got "bad_verification_code" on token exchange`)
	Equal(t, "", tokenInfo.AccessToken)

	testReturnCode = 200
	testResponseJson = `{"foo": "bar"}`
	tokenInfo, err = Authenticate(testConfigCopy, request)
	Error(t, err)
	EqualError(t, err, `error: no access_token on token exchange`)
	Equal(t, "", tokenInfo.AccessToken)

}

func Test_Authentication_ProviderError(t *testing.T) {
	request, _ := http.NewRequest("GET", testConfig.RedirectURI, nil)
	request.URL, _ = url.Parse("http://localhost/callback?error=provider_login_error")

	_, err := Authenticate(testConfig, request)

	Error(t, err)
	Equal(t, "error: provider_login_error", err.Error())
}

func Test_Authentication_StateError(t *testing.T) {
	request, _ := http.NewRequest("GET", testConfig.RedirectURI, nil)
	request.Header.Set("Cookie", "oauthState=XXXXXXX")
	request.URL, _ = url.Parse("http://localhost/callback?code=theCode&state=theState")

	_, err := Authenticate(testConfig, request)

	Error(t, err)
	Equal(t, "error: oauth state param could not be verified", err.Error())
}

func Test_Authentication_NoCodeError(t *testing.T) {
	request, _ := http.NewRequest("GET", testConfig.RedirectURI, nil)
	request.Header.Set("Cookie", "oauthState=theState")
	request.URL, _ = url.Parse("http://localhost/callback?state=theState")

	_, err := Authenticate(testConfig, request)

	Error(t, err)
	Equal(t, "error: no auth code provided", err.Error())
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

	Error(t, err)
	Equal(t, "error: expected http status 200 on token exchange, but got 500", err.Error())
}

func Test_Authentication_ProviderNetworkError(t *testing.T) {

	testConfigCopy := testConfig
	testConfigCopy.TokenURL = "http://localhost:12345678"

	request, _ := http.NewRequest("GET", testConfig.RedirectURI, nil)
	request.Header.Set("Cookie", "oauthState=theState")
	request.URL, _ = url.Parse("http://localhost/callback?code=theCode&state=theState")

	_, err := Authenticate(testConfigCopy, request)

	Error(t, err)
	Contains(t, err.Error(), "invalid port")
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

	Error(t, err)
	Equal(t, "error on parsing oauth token: unexpected end of JSON input", err.Error())
}
