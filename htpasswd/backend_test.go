package htpasswd

import (
	. "github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/login"
	"testing"
)

func TestSetup(t *testing.T) {
	p, exist := login.GetProvider(ProviderName)
	True(t, exist)
	NotNil(t, p)

	file := writeTmpfile(testfile)
	backend, err := p(map[string]string{
		"file": file,
	})

	NoError(t, err)
	Equal(t,
		file,
		backend.(*Backend).auth.filename)
}

func TestSetup_Error(t *testing.T) {
	p, exist := login.GetProvider(ProviderName)
	True(t, exist)
	NotNil(t, p)

	_, err := p(map[string]string{})
	Error(t, err)
}

func TestSimpleBackend_Authenticate(t *testing.T) {
	backend, err := NewBackend(writeTmpfile(testfile))
	NoError(t, err)

	authenticated, userInfo, err := backend.Authenticate("bob-bcrypt", "secret")
	True(t, authenticated)
	Equal(t, "bob-bcrypt", userInfo.Sub)
	NoError(t, err)

	authenticated, userInfo, err = backend.Authenticate("bob-bcrypt", "fooo")
	False(t, authenticated)
	Equal(t, "", userInfo.Sub)
	NoError(t, err)

	authenticated, userInfo, err = backend.Authenticate("", "")
	False(t, authenticated)
	Equal(t, "", userInfo.Sub)
	NoError(t, err)
}
