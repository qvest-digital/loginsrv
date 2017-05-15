package login

import (
	. "github.com/stretchr/testify/assert"
	"testing"
)

func TestSetup(t *testing.T) {
	p, exist := GetProvider(SimpleProviderName)
	True(t, exist)
	NotNil(t, p)

	backend, err := p(map[string]string{
		"bob": "secret",
	})

	NoError(t, err)
	Equal(t,
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
	True(t, authenticated)
	Equal(t, "bob", userInfo.Sub)
	NoError(t, err)

	authenticated, userInfo, err = backend.Authenticate("bob", "fooo")
	False(t, authenticated)
	Equal(t, "", userInfo.Sub)
	NoError(t, err)

	authenticated, userInfo, err = backend.Authenticate("", "")
	False(t, authenticated)
	Equal(t, "", userInfo.Sub)
	NoError(t, err)
}
