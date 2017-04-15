package oauth2

import (
	"fmt"
	"github.com/stretchr/testify/assert"
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
	RedirectURL:  "http://localhost/callback",
	Scopes:       []string{"email", "other"},
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
		url.QueryEscape(testConfig.RedirectURL),
		"email+other",
		state,
	)

	assert.Equal(t, expectedLocation, resp.Header().Get("Location"))
}
