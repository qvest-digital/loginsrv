package osiam

import (
	"errors"
	"fmt"
	"github.com/tarent/loginsrv/login"
	"net/url"
)

type OsiamBackend struct {
	endpoint     string
	clientId     string
	clientSecret string
}

// NewOsiamBackend creates a new OSIAM Backend and verifies the parameters.
func NewOsiamBackend(endpoint, clientId, clientSecret string) (*OsiamBackend, error) {
	if _, err := url.Parse(endpoint); err != nil {
		return nil, fmt.Errorf("osiam endpoint hast to be a valid url: %v: %v", endpoint, err)
	}

	if clientId == "" {
		return nil, errors.New("No osiam clientId provided.")
	}
	if clientSecret == "" {
		return nil, errors.New("No osiam clientSecret provided.")
	}
	return &OsiamBackend{
		endpoint:     endpoint,
		clientId:     clientId,
		clientSecret: clientSecret,
	}, nil
}

func (ob *OsiamBackend) Authenticate(username, password string) (bool, login.UserInfo, error) {
	return false, login.UserInfo{}, errors.New("Not implemented yet")
}
