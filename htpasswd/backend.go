package htpasswd

import (
	"errors"
	"github.com/tarent/loginsrv/login"
)

const ProviderName = "htpasswd"

func init() {
	login.RegisterProvider(
		&login.ProviderDescription{
			Name: ProviderName,
		},
		BackendFactory)
}

func BackendFactory(config map[string]string) (login.Backend, error) {
	if f, exist := config["file"]; exist {
		return NewBackend(f)
	}
	return nil, errors.New(`missing parameter "file" for htpasswd provider.`)
}

// Backend is a htpasswd based authentication backend.
type Backend struct {
	auth *Auth
}

// NewBackend creates a new  Backend and verifies the parameters.
func NewBackend(filename string) (*Backend, error) {
	auth, err := NewAuth(filename)
	return &Backend{
		auth,
	}, err
}

func (sb *Backend) Authenticate(username, password string) (bool, login.UserInfo, error) {
	authenticated, err := sb.auth.Authenticate(username, password)
	if authenticated && err == nil {
		return authenticated, login.UserInfo{Username: username}, err
	}
	return false, login.UserInfo{}, err
}
