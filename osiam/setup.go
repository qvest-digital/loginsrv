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
		func(config map[string]string) (login.Backend, error) {
			return NewBackend(config["endpoint"], config["clientId"], config["clientSecret"])
		})
}
