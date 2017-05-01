package login

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetup(t *testing.T) {
	p, exist := GetProvider(SimpleProviderName)
	assert.True(t, exist)
	assert.NotNil(t, p)

	backend, err := p(map[string]string{
		"bob": "secret",
	})

	assert.NoError(t, err)
	assert.Equal(t,
		map[string]string{
			"bob": "secret",
		},
		backend.(*SimpleBackend).userPassword)
}

func TestSimpleBackend_Authenticate(t *testing.T) {
	backend := NewSimpleBackend(map[string]string{
		"bob": "secret",
	})

	authenticated, userInfo, err := backend.Authenticate("bob", "secret")
	assert.True(t, authenticated)
	assert.Equal(t, "bob", userInfo.Sub)
	assert.NoError(t, err)

	authenticated, userInfo, err = backend.Authenticate("bob", "fooo")
	assert.False(t, authenticated)
	assert.Equal(t, "", userInfo.Sub)
	assert.NoError(t, err)

	authenticated, userInfo, err = backend.Authenticate("", "")
	assert.False(t, authenticated)
	assert.Equal(t, "", userInfo.Sub)
	assert.NoError(t, err)
}
