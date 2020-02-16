package saml

// Provider is the description of a SAML provider adapter
type Provider struct {
	Name string
}

var provider = map[string]Provider{}

// RegisterProvider a SAML provider
func RegisterProvider(p Provider) {
	provider[p.Name] = p
}

// UnRegisterProvider removes a provider
func UnRegisterProvider(name string) {
	delete(provider, name)
}

// GetProvider returns a provider
func GetProvider(providerName string) (Provider, bool) {
	p, exist := provider[providerName]
	return p, exist
}

// ProviderList returns the names of all registered SAML providers
func ProviderList() []string {
	list := make([]string, 0, len(provider))
	for k := range provider {
		list = append(list, k)
	}
	return list
}
