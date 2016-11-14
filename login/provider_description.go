package login

type ProviderDescription struct {
	// the name of the provider
	Name string

	// the config options, which the provider supports
	options []ProviderOption
}

type ProviderOption struct {
	Name        string
	Description string
	Default     string
	Required    string
}
