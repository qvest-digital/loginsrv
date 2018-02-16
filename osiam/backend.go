package osiam

import (
	"errors"
	"fmt"
	"github.com/tarent/loginsrv/model"
	"net/url"
)

// Backend is the osiam authentication backend.
type Backend struct {
	client *Client
}

// NewBackend creates a new OSIAM Backend and verifies the parameters.
func NewBackend(endpoint, clientID, clientSecret string) (*Backend, error) {
	if _, err := url.Parse(endpoint); err != nil {
		return nil, fmt.Errorf("osiam endpoint has to be a valid url: %v: %v", endpoint, err)
	}

	if clientID == "" {
		return nil, errors.New("no osiam clientID provided.")
	}
	if clientSecret == "" {
		return nil, errors.New("no osiam clientSecret provided")
	}
	client := NewClient(endpoint, clientID, clientSecret)
	return &Backend{
		client: client,
	}, nil
}

// Authenticate the user
func (b *Backend) Authenticate(username, password string) (bool, model.UserInfo, error) {
	authenticated, _, err := b.client.GetTokenByPassword(username, password)
	if !authenticated || err != nil {
		return authenticated, model.UserInfo{}, err
	}
	userInfo := model.UserInfo{
		Origin: OsiamProviderName,
		Sub:    username,
	}
	return true, userInfo, nil
}
