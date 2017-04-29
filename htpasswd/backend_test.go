package htpasswd

import (
	"github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/login"
	"testing"
)

func TestSetup(t *testing.T) {
	p, exist := login.GetProvider(ProviderName)
	assert.True(t, exist)
	assert.NotNil(t, p)

	file := writeTmpfile(testfile)
	backend, err := p(map[string]string{
		"file": file,
	})

	assert.NoError(t, err)
	assert.Equal(t,
		file,
		backend.(*Backend).auth.filename)
}

func TestSetup_Error(t *testing.T) {
	p, exist := login.GetProvider(ProviderName)
	assert.True(t, exist)
	assert.NotNil(t, p)

	_, err := p(map[string]string{})
	assert.Error(t, err)
}

func TestSimpleBackend_Authenticate(t *testing.T) {
	backend, err := NewBackend(writeTmpfile(testfile))
	assert.NoError(t, err)

	authenticated, userInfo, err := backend.Authenticate("bob-bcrypt", "secret")
	assert.True(t, authenticated)
	assert.Equal(t, "bob-bcrypt", userInfo.Sub)
	assert.NoError(t, err)

	authenticated, userInfo, err = backend.Authenticate("bob-bcrypt", "fooo")
	assert.False(t, authenticated)
	assert.Equal(t, "", userInfo.Sub)
	assert.NoError(t, err)

	authenticated, userInfo, err = backend.Authenticate("", "")
	assert.False(t, authenticated)
	assert.Equal(t, "", userInfo.Sub)
	assert.NoError(t, err)
}
