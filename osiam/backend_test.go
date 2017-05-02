package osiam

import (
	"github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/model"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBackend_Authenticate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(osiamMockHandler))
	defer server.Close()

	// positive case
	backend, err := NewBackend(server.URL, "example-client", "secret")
	assert.NoError(t, err)
	authenticated, userInfo, err := backend.Authenticate("admin", "koala")

	assert.NoError(t, err)
	assert.True(t, authenticated)
	assert.Equal(t,
		model.UserInfo{
			Sub: "admin",
		},
		userInfo)

	// wrong client credentials
	backend, err = NewBackend(server.URL, "example-client", "XXX")
	assert.NoError(t, err)
	authenticated, userInfo, err = backend.Authenticate("admin", "koala")
	assert.Error(t, err)
	assert.False(t, authenticated)

	// wrong user credentials
	backend, err = NewBackend(server.URL, "example-client", "secret")
	assert.NoError(t, err)
	authenticated, userInfo, err = backend.Authenticate("admin", "XXX")
	assert.NoError(t, err)
	assert.False(t, authenticated)

}

func TestBackend_AuthenticateErrorCases(t *testing.T) {
	_, err := NewBackend("://", "example-client", "secret")
	assert.Error(t, err)

	_, err = NewBackend("http://example.com", "", "secret")
	assert.Error(t, err)

	_, err = NewBackend("http://example.com", "example-client", "")
	assert.Error(t, err)
}
