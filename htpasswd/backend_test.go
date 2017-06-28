package htpasswd

import (
	. "github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/login"
	"strings"
	"testing"
)

func TestSetupOneFile(t *testing.T) {
	p, exist := login.GetProvider(ProviderName)
	True(t, exist)
	NotNil(t, p)

	file := writeTmpfile(testfile)
	backend, err := p(map[string]string{
		"file": file,
	})

	NoError(t, err)
	Equal(t,
		[]File{File{name: file}},
		backend.(*Backend).auth.filenames)
}

func TestSetupTwoFiles(t *testing.T) {
	p, exist := login.GetProvider(ProviderName)
	True(t, exist)
	NotNil(t, p)

	filenames := []string{writeTmpfile(testfile), writeTmpfile(testfile)}

	var morphed []File
	for _, curFile := range filenames {
		morphed = append(morphed, File{name: curFile})
	}
	backend, err := p(map[string]string{
		"file": strings.Join(filenames, ","),
	})

	NoError(t, err)
	Equal(t,
		morphed,
		backend.(*Backend).auth.filenames)
}

func TestSetupTwoConfigs(t *testing.T) {
	p, exist := login.GetProvider(ProviderName)
	True(t, exist)
	NotNil(t, p)

	configFiles := []string{writeTmpfile(testfile), writeTmpfile(testfile)}
	configFile := []string{writeTmpfile(testfile), writeTmpfile(testfile)}

	var morphed []File
	for _, curFile := range append(configFiles, configFile...) {
		morphed = append(morphed, File{name: curFile})
	}

	backend, err := p(map[string]string{
		"files": strings.Join(configFiles, ","),
		"file":  strings.Join(configFile, ","),
	})

	NoError(t, err)
	Equal(t,
		morphed,
		backend.(*Backend).auth.filenames)
}

func TestSetup_Error(t *testing.T) {
	p, exist := login.GetProvider(ProviderName)
	True(t, exist)
	NotNil(t, p)

	_, err := p(map[string]string{})
	Error(t, err)
}

func TestSimpleBackend_Authenticate(t *testing.T) {
	backend, err := NewBackend([]string{writeTmpfile(testfile)})
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
