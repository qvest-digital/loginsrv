package osiam

import (
	"github.com/tarent/loginsrv/login"
)

const OsiamProviderName = "osiam"

func init() {
	login.RegisterProvider(
		&login.ProviderDescription{
			Name: OsiamProviderName,
		},
		OsiamBackendFactory)
}

func OsiamBackendFactory(config map[string]string) (login.Backend, error) {
	return NewOsiamBackend(config["endpoint"], config["clientId"], config["clientSecret"])
}
