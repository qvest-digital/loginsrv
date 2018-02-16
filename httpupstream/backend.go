package httpupstream

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/tarent/loginsrv/login"
	"github.com/tarent/loginsrv/model"
)

// ProviderName const
const ProviderName = "httpupstream"

const defaultTimeout = time.Minute

func init() {
	login.RegisterProvider(
		&login.ProviderDescription{
			Name:     ProviderName,
			HelpText: "Httpupstream login backend opts: upstream=...,skipverify=...,timeout=...",
		},
		BackendFactory)
}

// BackendFactory creates a httpupstream backend
func BackendFactory(config map[string]string) (login.Backend, error) {
	us, ue := config["upstream"]
	ts, te := config["timeout"]
	vs, ve := config["skipverify"]

	if !ue {
		return nil, errors.New(`missing parameter "upstream" for httpupstream provider`)
	}

	u, err := url.Parse(us)
	if err != nil {
		return nil, fmt.Errorf(`invalid parameter value "%s" in "upstream" httpupstream provider: %v`, us, err)
	}

	v := false
	t := defaultTimeout

	if te {
		t, err = time.ParseDuration(ts)
		if err != nil {
			return nil, fmt.Errorf(`invalid parameter value "%s" in "timeout" httpupstream provider: %v`, ts, err)
		}
	}

	if ve {
		v, err = strconv.ParseBool(vs)
		if err != nil {
			return nil, fmt.Errorf(`invalid parameter value "%s" in "skipverify" httpupstream provider: %v`, ts, err)
		}
	}

	return NewBackend(u, t, v)
}

// Backend is a httpupstream based authentication backend.
type Backend struct {
	auth *Auth
}

// NewBackend creates a new Backend and verifies the parameters.
func NewBackend(upstream *url.URL, timeout time.Duration, skipverify bool) (*Backend, error) {
	auth, err := NewAuth(upstream, timeout, skipverify)
	return &Backend{
		auth,
	}, err
}

// Authenticate the user
func (sb *Backend) Authenticate(username, password string) (bool, model.UserInfo, error) {
	authenticated, err := sb.auth.Authenticate(username, password)
	if authenticated && err == nil {
		return authenticated, model.UserInfo{
			Origin: ProviderName,
			Sub:    username,
		}, err
	}
	return false, model.UserInfo{}, err
}
