package oauth2

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_Manager_Positive_Flow(t *testing.T) {
	var startFlowCalled, authenticateCalled, getUserInfoCalled bool
	var startFlowReceivedConfig, authenticateReceivedConfig Config
	expectedToken := TokenInfo{AccessToken: "the-access-token"}

	exampleProvider := Provider{
		Name:     "example",
		AuthURL:  "https://example.com/login/oauth/authorize",
		TokenURL: "https://example.com/login/oauth/access_token",
		GetUserInfo: func(token TokenInfo) (map[string]string, error) {
			getUserInfoCalled = true
			assert.Equal(t, token, expectedToken)
			return map[string]string{
				"username": "the-username",
			}, nil
		},
	}
	RegisterProvider(exampleProvider)
	defer UnRegisterProvider(exampleProvider.Name)

	expectedConfig := Config{
		ClientID:     "client42",
		ClientSecret: "secret",
		AuthURL:      exampleProvider.AuthURL,
		TokenURL:     exampleProvider.TokenURL,
		RedirectURI:  "http://localhost/callback",
		Scope:        "email other",
		Provider:     exampleProvider,
	}

	m := NewManager()
	m.AddConfig(exampleProvider.Name, map[string]string{
		"client_id":     expectedConfig.ClientID,
		"client_secret": expectedConfig.ClientSecret,
		"scope":         expectedConfig.Scope,
		"redirect_uri":  expectedConfig.RedirectURI,
	})

	m.startFlow = func(cfg Config, w http.ResponseWriter) {
		startFlowCalled = true
		startFlowReceivedConfig = cfg
	}

	m.authenticate = func(cfg Config, r *http.Request) (TokenInfo, error) {
		authenticateCalled = true
		authenticateReceivedConfig = cfg
		return expectedToken, nil
	}

	// start flow
	r, _ := http.NewRequest("GET", "http://example.com/login/"+exampleProvider.Name, nil)

	startedFlow, authenticated, userInfo, err := m.Handle(httptest.NewRecorder(), r)
	assert.NoError(t, err)
	assert.True(t, startedFlow)
	assert.False(t, authenticated)
	assert.Equal(t, UserInfo{}, userInfo)

	assert.True(t, startFlowCalled)
	assert.False(t, authenticateCalled)

	assertEqualConfig(t, expectedConfig, startFlowReceivedConfig)

	// callback
	r, _ = http.NewRequest("GET", "http://example.com/login/"+exampleProvider.Name+callbackPathSuffix, nil)

	startedFlow, authenticated, userInfo, err = m.Handle(httptest.NewRecorder(), r)
	assert.NoError(t, err)
	assert.False(t, startedFlow)
	assert.True(t, authenticated)
	assert.Equal(t, UserInfo{Username: "the-username"}, userInfo)
	assert.True(t, authenticateCalled)
	assertEqualConfig(t, expectedConfig, authenticateReceivedConfig)

	assert.True(t, getUserInfoCalled)
}

func assertEqualConfig(t *testing.T, c1, c2 Config) {
	assert.Equal(t, c1.AuthURL, c2.AuthURL)
	assert.Equal(t, c1.ClientID, c2.ClientID)
	assert.Equal(t, c1.ClientSecret, c2.ClientSecret)
	assert.Equal(t, c1.Scope, c2.Scope)
	assert.Equal(t, c1.RedirectURI, c2.RedirectURI)
	assert.Equal(t, c1.TokenURL, c2.TokenURL)
	assert.Equal(t, c1.Provider.Name, c2.Provider.Name)
}
