package osiam

import (
	"errors"
	"fmt"
	"github.com/tarent/loginsrv/login"
	"net/url"
)

type Backend struct {
	client *Client
}

// NewBackend creates a new OSIAM Backend and verifies the parameters.
func NewBackend(endpoint, clientId, clientSecret string) (*Backend, error) {
	if _, err := url.Parse(endpoint); err != nil {
		return nil, fmt.Errorf("osiam endpoint has to be a valid url: %v: %v", endpoint, err)
	}

	if clientId == "" {
		return nil, errors.New("No osiam clientId provided.")
	}
	if clientSecret == "" {
		return nil, errors.New("No osiam clientSecret provided.")
	}
	client := NewClient(endpoint, clientId, clientSecret)
	return &Backend{
		client: client,
	}, nil
}

func (b *Backend) Authenticate(username, password string) (bool, login.UserInfo, error) {
	authenticated, _, err := b.client.GetTokenByPassword(username, password)
	if !authenticated || err != nil {
		return authenticated, login.UserInfo{}, err
	}
	userInfo := login.UserInfo{
		Username: username,
	}
	return true, userInfo, nil
}
