package oauth2

import (
	"github.com/tarent/loginsrv/model"
)

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
	GetUserInfo func(token TokenInfo) (u model.UserInfo, rawUserJson string, err error)
}

var provider = map[string]Provider{}

// RegisterProvider an Oauth provider
func RegisterProvider(p Provider) {
	provider[p.Name] = p
}

// Unregister an Oauth provider
func UnRegisterProvider(name string) {
	delete(provider, name)
}

// GetProvider returns a provider
func GetProvider(providerName string) (Provider, bool) {
	p, exist := provider[providerName]
	return p, exist
}

// ProviderList returns the names of all registered provider
func ProviderList() []string {
	list := make([]string, 0, len(provider))
	for k := range provider {
		list = append(list, k)
	}
	return list
}
