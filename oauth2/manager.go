package oauth2

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const callbackPathSuffix = "/callback"

// Oauth provider configuration
type Provider struct {
	// The name to access the provider in the configuration
	Name string

	// The oauth authentication url to redirect to
	AuthURL string

	// The url for token exchange
	TokenURL string

	// GetUserInfo is a provider specific Implementation
	// for fetching the user information.
	// Possible keys in the returned map are:
	// username, email, name
	GetUserInfo func(token TokenInfo) (map[string]string, error)
}

var DefaultManager = NewManager()

// The manager has the responsibility to pick the right configuration
// and start the oauth habdling.
type Manager struct {
	provider     map[string]Provider
	configs      map[string]Config
	startFlow    func(cfg Config, w http.ResponseWriter)
	authenticate func(cfg Config, r *http.Request) (TokenInfo, error)
}

// NewManager creates a new Manager
func NewManager() *Manager {
	return &Manager{
		provider:     map[string]Provider{},
		configs:      map[string]Config{},
		startFlow:    StartFlow,
		authenticate: Authenticate,
	}
}

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
		userInfo = UserInfo{
			Username: tokenInfo.AccessToken,
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

	if cfg.RedirectURL == "" {
		cfg.RedirectURL = redirectUriFromRequest(r)
	}

	return cfg, nil
}

func (manager *Manager) getConfigNameFromPath(path string) string {
	path = strings.TrimSuffix(path, "/"+callbackPathSuffix)
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

// Register an Oauth provider
func (manager *Manager) RegisterProvider(provider Provider) {
	manager.provider[provider.Name] = provider
}

// ProviderList returns the names of all registered providre
func (manager *Manager) ProviderList() []string {
	list := make([]string, 0, len(manager.provider))
	for k, _ := range manager.provider {
		list = append(list, k)
	}
	return list
}

// Add a configuration for a provider
func (manager *Manager) AddConfig(providerName string, opts map[string]string) error {
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
