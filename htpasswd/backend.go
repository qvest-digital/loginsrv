package htpasswd

import (
	"errors"
	"github.com/tarent/loginsrv/login"
	"github.com/tarent/loginsrv/model"
	"strings"
)

// ProviderName const
const ProviderName = "htpasswd"

func init() {
	login.RegisterProvider(
		&login.ProviderDescription{
			Name:     ProviderName,
			HelpText: "Htpasswd login backend opts: files=/path/to/pwdfile,/path/to/additionalfile",
		},
		BackendFactory)
}

// BackendFactory creates a htpasswd backend
func BackendFactory(config map[string]string) (login.Backend, error) {
	var files []string

	if f, exist := config["files"]; exist {
		for _, file := range strings.Split(f, ",") {
			files = append(files, file)
		}
	}

	if f, exist := config["file"]; exist {
		for _, file := range strings.Split(f, ",") {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		return nil, errors.New(`missing parameter "file" for htpasswd provider`)
	}

	return NewBackend(files)
}

// Backend is a htpasswd based authentication backend.
type Backend struct {
	auth *Auth
}

// NewBackend creates a new Backend and verifies the parameters.
func NewBackend(filenames []string) (*Backend, error) {
	auth, err := NewAuth(filenames)
	return &Backend{
		auth,
	}, err
}

// Authenticate the user
func (sb *Backend) Authenticate(username, password string) (bool, model.UserInfo, error) {
	authenticated, err := sb.auth.Authenticate(username, password)
	if authenticated && err == nil {
		return authenticated, model.UserInfo{Sub: username}, err
	}
	return false, model.UserInfo{}, err
}
