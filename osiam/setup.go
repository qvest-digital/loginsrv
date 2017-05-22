package osiam

import (
	"github.com/tarent/loginsrv/logging"
	"github.com/tarent/loginsrv/login"
)

// OsiamProviderName const with the name of the provider
const OsiamProviderName = "osiam"

func init() {
	login.RegisterProvider(
		&login.ProviderDescription{
			Name:     OsiamProviderName,
			HelpText: "Osiam login backend opts: endpoint=..,client_id=..,client_secret=..",
		},
		func(config map[string]string) (login.Backend, error) {
			if config["clientId"] != "" {
				logging.Logger.Warn("DEPRECATED: please use 'client_id' and 'client_secret' in future.")
				return NewBackend(config["endpoint"], config["clientId"], config["clientSecret"])
			}
			return NewBackend(config["endpoint"], config["client_id"], config["client_secret"])
		})
}
