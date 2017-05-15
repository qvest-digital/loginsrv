package oauth2

import (
	"crypto/tls"
	"errors"
	. "github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/model"
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
		GetUserInfo: func(token TokenInfo) (model.UserInfo, string, error) {
			getUserInfoCalled = true
			Equal(t, token, expectedToken)
			return model.UserInfo{
				Sub: "the-username",
			}, "", nil
		},
	}
	RegisterProvider(exampleProvider)
	defer UnRegisterProvider(exampleProvider.Name)

	expectedConfig := Config{
		ClientID:     "client42",
		ClientSecret: "secret",
		AuthURL:      exampleProvider.AuthURL,
		TokenURL:     exampleProvider.TokenURL,
		RedirectURI:  "http://localhost",
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
	NoError(t, err)
	True(t, startedFlow)
	False(t, authenticated)
	Equal(t, model.UserInfo{}, userInfo)

	True(t, startFlowCalled)
	False(t, authenticateCalled)

	assertEqualConfig(t, expectedConfig, startFlowReceivedConfig)

	// callback
	r, _ = http.NewRequest("GET", "http://example.com/login/"+exampleProvider.Name+"?code=xyz", nil)

	startedFlow, authenticated, userInfo, err = m.Handle(httptest.NewRecorder(), r)
	NoError(t, err)
	False(t, startedFlow)
	True(t, authenticated)
	Equal(t, model.UserInfo{Sub: "the-username"}, userInfo)
	True(t, authenticateCalled)
	assertEqualConfig(t, expectedConfig, authenticateReceivedConfig)

	True(t, getUserInfoCalled)
}

func Test_Manager_NoAauthOnWrongCode(t *testing.T) {
	var authenticateCalled, getUserInfoCalled bool

	exampleProvider := Provider{
		Name:     "example",
		AuthURL:  "https://example.com/login/oauth/authorize",
		TokenURL: "https://example.com/login/oauth/access_token",
		GetUserInfo: func(token TokenInfo) (model.UserInfo, string, error) {
			getUserInfoCalled = true
			return model.UserInfo{}, "", nil
		},
	}
	RegisterProvider(exampleProvider)
	defer UnRegisterProvider(exampleProvider.Name)

	m := NewManager()
	m.AddConfig(exampleProvider.Name, map[string]string{
		"client_id":     "foo",
		"client_secret": "bar",
	})

	m.authenticate = func(cfg Config, r *http.Request) (TokenInfo, error) {
		authenticateCalled = true
		return TokenInfo{}, errors.New("code not valid")
	}

	// callback
	r, _ := http.NewRequest("GET", "http://example.com/login/"+exampleProvider.Name+"?code=xyz", nil)

	startedFlow, authenticated, userInfo, err := m.Handle(httptest.NewRecorder(), r)
	EqualError(t, err, "code not valid")
	False(t, startedFlow)
	False(t, authenticated)
	Equal(t, model.UserInfo{}, userInfo)
	True(t, authenticateCalled)
	False(t, getUserInfoCalled)
}

func Test_Manager_getConfig_ErrorCase(t *testing.T) {
	r, _ := http.NewRequest("GET", "http://example.com/login", nil)

	m := NewManager()
	m.AddConfig("github", map[string]string{
		"client_id":     "foo",
		"client_secret": "bar",
	})

	_, err := m.GetConfigFromRequest(r)
	EqualError(t, err, "no oauth configuration for login")
}

func Test_Manager_AddConfig_ErrorCases(t *testing.T) {
	m := NewManager()

	NoError(t,
		m.AddConfig("github", map[string]string{
			"client_id":     "foo",
			"client_secret": "bar",
		}))

	EqualError(t,
		m.AddConfig("FOOOO", map[string]string{
			"client_id":     "foo",
			"client_secret": "bar",
		}),
		"no provider for name FOOOO",
	)

	EqualError(t,
		m.AddConfig("github", map[string]string{
			"client_secret": "bar",
		}),
		"missing parameter client_id",
	)

	EqualError(t,
		m.AddConfig("github", map[string]string{
			"client_id": "foo",
		}),
		"missing parameter client_secret",
	)

}

func Test_Manager_redirectUriFromRequest(t *testing.T) {
	tests := []struct {
		url      string
		tls      bool
		header   http.Header
		expected string
	}{
		{
			"http://example.com/login/github",
			false,
			http.Header{},
			"http://example.com/login/github",
		},
		{
			"http://localhost/login/github",
			false,
			http.Header{
				"X-Forwarded-Host": {"example.com"},
			},
			"http://example.com/login/github",
		},
		{
			"http://localhost/login/github",
			true,
			http.Header{
				"X-Forwarded-Host": {"example.com"},
			},
			"https://example.com/login/github",
		},
		{
			"http://localhost/login/github",
			false,
			http.Header{
				"X-Forwarded-Host":  {"example.com"},
				"X-Forwarded-Proto": {"https"},
			},
			"https://example.com/login/github",
		},
	}
	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			r, _ := http.NewRequest("GET", test.url, nil)
			r.Header = test.header
			if test.tls {
				r.TLS = &tls.ConnectionState{}
			}
			uri := redirectURIFromRequest(r)
			Equal(t, test.expected, uri)
		})
	}
}

func Test_Manager_RedirectURI_Generation(t *testing.T) {
	var startFlowReceivedConfig Config

	m := NewManager()
	m.AddConfig("github", map[string]string{
		"client_id":     "foo",
		"client_secret": "bar",
		"scope":         "bazz",
	})

	m.startFlow = func(cfg Config, w http.ResponseWriter) {
		startFlowReceivedConfig = cfg
	}

	callURL := "http://example.com/login/github"
	r, _ := http.NewRequest("GET", callURL, nil)

	_, _, _, err := m.Handle(httptest.NewRecorder(), r)
	NoError(t, err)
	Equal(t, callURL, startFlowReceivedConfig.RedirectURI)
}

func assertEqualConfig(t *testing.T, c1, c2 Config) {
	Equal(t, c1.AuthURL, c2.AuthURL)
	Equal(t, c1.ClientID, c2.ClientID)
	Equal(t, c1.ClientSecret, c2.ClientSecret)
	Equal(t, c1.Scope, c2.Scope)
	Equal(t, c1.RedirectURI, c2.RedirectURI)
	Equal(t, c1.TokenURL, c2.TokenURL)
	Equal(t, c1.Provider.Name, c2.Provider.Name)
}
