package login

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tarent/loginsrv/model"
	"github.com/tarent/loginsrv/oauth2"
)

func TestVHostSelection(t *testing.T) {
	testCases := []struct {
		hostname string
		username string
		password string
		expect   bool
	}{
		{"default.domain.org", "bob", "secret", true},
		{"default.domain.org", "bob", "admin", false},
		{"foo.domain.org", "bob", "thebuilder", true},
		{"foo.domain.org", "bob", "secret", false},
		{"bar.domain.org", "bob", "thebuilder", false},
		{"bar.domain.org", "bob", "admin", true},
	}

	var configFile = `
    vhosts:
      - name: foo
        backends:
          - name: simple
            bob: thebuilder
      - name: bar
        backends:
           - name: simple
             bob: admin
    `

	config := DefaultConfig()
	config.Backends = Options{
		SimpleProviderName: map[string]string{
			"bob": "secret",
		},
	}
	require.NoError(t, parseConfigData(config, []byte(configFile)))

	handler, err := NewHandler(config)
	require.NoError(t, err)

	for i, testCase := range testCases {
		data := struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{testCase.username, testCase.password}
		body, err := json.Marshal(&data)
		assert.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost, config.LoginPath, bytes.NewReader(body))
		req.URL.Host = testCase.hostname
		req.Header.Set("Content-type", "application/json")
		require.NoError(t, err)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, testCase.expect, rec.Code == http.StatusOK, fmt.Sprintf("test case %d", i))
	}
}

func TestVHostOAuth(t *testing.T) {
	var configFile = `
    vhosts:
      - name: foo
        oauth:
          - provider: mock-vhost-oauth
            client-id: me@example.com
            client-secret: secrect
    `

	config := DefaultConfig()
	require.NoError(t, parseConfigData(config, []byte(configFile)))
	handler, err := NewHandler(config)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, config.LoginPath+"/mock-vhost-oauth", nil)
	req.URL.Host = "foo.example.org"
	req.Header.Set("Content-type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusFound, rec.Code)
	loc, err := url.Parse(rec.Header().Get("Location"))
	assert.NoError(t, err)
	assert.Equal(t, "mock-auth-service", loc.Hostname())
}

func makeMockVHostOAuthProvier() oauth2.Provider {
	return oauth2.Provider{
		Name:     "mock-vhost-oauth",
		AuthURL:  "https://mock-auth-service/auth",
		TokenURL: "https://mock-auth-service/token",
		GetUserInfo: func(token oauth2.TokenInfo) (model.UserInfo, string, error) {
			return model.UserInfo{Sub: "marvin"}, "", nil
		},
	}
}

func TestMain(m *testing.M) {
	oauth2.RegisterProvider(makeMockVHostOAuthProvier())
	os.Exit(m.Run())
}
