package oauth2

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const callbackPathSuffix = "/callback"

// The manager has the responsibility to handle the user user requests in an oauth flow.
// It has to pick the right configuration and start the oauth redirecting.
type Manager struct {
	configs      map[string]Config
	startFlow    func(cfg Config, w http.ResponseWriter)
	authenticate func(cfg Config, r *http.Request) (TokenInfo, error)
}

// NewManager creates a new Manager
func NewManager() *Manager {
	return &Manager{
		configs:      map[string]Config{},
		startFlow:    StartFlow,
		authenticate: Authenticate,
	}
}

// Handle is managing the oauth flow.
// Dependent on the suffix of the url, the oauth flow is started or
// the call is interpreted as the redirect callback and the token exchange is done.
// Return parameters:
//   startedFlow - true, if this was the initial call to start the oauth flow
//   authenticated - if the authentication was successful or not
//   userInfo - the user info from the provider in case of a succesful authentication
//   err - an error
func (manager *Manager) Handle(w http.ResponseWriter, r *http.Request) (
	startedFlow bool,
	authenticated bool,
	userInfo UserInfo,
	err error) {

	cfg, err := manager.getConfig(r)
	if err != nil {
		return false, false, UserInfo{}, err
	}

	if strings.HasSuffix(r.URL.Path, callbackPathSuffix) {
		tokenInfo, err := manager.authenticate(cfg, r)
		if err != nil {
			return false, false, UserInfo{}, err
		}

		ui, err := cfg.Provider.GetUserInfo(tokenInfo)
		if err != nil {
			return false, false, UserInfo{}, err
		}
		userInfo = UserInfo{
			Username: ui["username"],
		}
		return false, true, userInfo, err
	}

	manager.startFlow(cfg, w)
	return true, false, UserInfo{}, nil
}

func (manager *Manager) getConfig(r *http.Request) (Config, error) {
	configName := manager.getConfigNameFromPath(r.URL.Path)
	cfg, exist := manager.configs[configName]
	if !exist {
		return Config{}, fmt.Errorf("no oauth configuration for %v", configName)
	}

	if cfg.RedirectURI == "" {
		cfg.RedirectURI = redirectUriFromRequest(r)
	}

	return cfg, nil
}

func (manager *Manager) getConfigNameFromPath(path string) string {
	path = strings.TrimSuffix(path, callbackPathSuffix)
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

// Add a configuration for a provider
func (manager *Manager) AddConfig(providerName string, opts map[string]string) error {
	p, exist := GetProvider(providerName)
	if !exist {
		return fmt.Errorf("no provider for name %v", providerName)
	}

	cfg := Config{
		Provider: p,
		AuthURL:  p.AuthURL,
		TokenURL: p.TokenURL,
	}

	if clientId, exist := opts["client_id"]; !exist {
		return fmt.Errorf("missing parameter client_id")
	} else {
		cfg.ClientID = clientId
	}

	if clientSecret, exist := opts["client_secret"]; !exist {
		return fmt.Errorf("missing parameter client_secret")
	} else {
		cfg.ClientSecret = clientSecret
	}

	if scope, exist := opts["scope"]; exist {
		cfg.Scope = scope
	}

	if redirectURI, exist := opts["redirect_uri"]; exist {
		cfg.RedirectURI = redirectURI
	}

	manager.configs[providerName] = cfg
	return nil
}

func redirectUriFromRequest(r *http.Request) string {
	u := url.URL{}
	u.Path = r.URL.Path + callbackPathSuffix

	if ffh := r.Header.Get("X-Forwarded-Host"); ffh == "" {
		u.Host = r.Host
	} else {
		u.Host = ffh
	}

	if ffp := r.Header.Get("X-Forwarded-Proto"); ffp == "" {
		if r.TLS != nil {
			u.Scheme = "https"
		} else {
			u.Scheme = "http"
		}
	} else {
		u.Scheme = ffp
	}

	return u.String()
}
