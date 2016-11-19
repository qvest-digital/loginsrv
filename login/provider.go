package login

// Provider is a factory method for creation of login backends.
type Provider func(config map[string]string) (Backend, error)

var provider = map[string]Provider{}
var providerDescription = map[string]*ProviderDescription{}

// RegisterProvider registers a factory method by the provider name.
func RegisterProvider(desc *ProviderDescription, factoryMethod Provider) {
	provider[desc.Name] = factoryMethod
	providerDescription[desc.Name] = desc
}

// GetProvider returns a registered provider by its name.
// The bool return parameter indicated, if there was such a provider.
func GetProvider(providerName string) (Provider, bool) {
	p, exist := provider[providerName]
	return p, exist
}
