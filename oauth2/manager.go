package oauth2

import ()

// Oauth provider configuration
type Provider interface {
	// The oauth authentication url to redirect to
	AuthURL() string

	// The url for token exchange
	TokenURL() string
}
type Manager struct {
	configs map[string]map[string]string
}

func NewManager() *Manager {
	return &Manager{
		configs: map[string]map[string]string{},
	}
}

func (oauth *Manager) AddConfig(opts map[string]string) error {
	return nil
}
