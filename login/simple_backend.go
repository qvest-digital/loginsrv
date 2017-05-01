package login

import (
	"errors"
	"github.com/tarent/loginsrv/model"
)

const SimpleProviderName = "simple"

func init() {
	RegisterProvider(
		&ProviderDescription{
			Name:     SimpleProviderName,
			HelpText: "Simple login backend opts: user1=password,user2=password,..",
		},
		SimpleBackendFactory)
}

func SimpleBackendFactory(config map[string]string) (Backend, error) {
	userPassword := map[string]string{}
	for k, v := range config {
		userPassword[k] = v
	}
	if len(userPassword) == 0 {
		return nil, errors.New("no users provided for simple backend")
	}
	return NewSimpleBackend(userPassword), nil
}

// Simple backend, working on a map of username password pairs
type SimpleBackend struct {
	userPassword map[string]string
}

// NewSimpleBackend creates a new SIMPLE Backend and verifies the parameters.
func NewSimpleBackend(userPassword map[string]string) *SimpleBackend {
	return &SimpleBackend{
		userPassword: userPassword,
	}
}

func (sb *SimpleBackend) Authenticate(username, password string) (bool, model.UserInfo, error) {
	if p, exist := sb.userPassword[username]; exist && p == password {
		return true, model.UserInfo{Sub: username}, nil
	}
	return false, model.UserInfo{}, nil
}
