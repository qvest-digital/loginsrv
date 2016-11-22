package htpasswd

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

// password for all of them is 'secret'
const testfile = `bob-md5:$apr1$IDZSCL/o$N68zaFDDRivjour94OVeB.

bob-bcrypt:$2y$05$Hw6y1sFwh6CdwiPOKFMYj..xVSQWI3wzyQvt5th392ig8RLmeLU.6
bob-sha:{SHA}5en6G6MezRroT3XKqkdPOmY/BfQ=

# a comment
bob-foo:{fooo}sdcsdcsdc/BfQ=


`

func TestClient_Hashes(t *testing.T) {
	auth, err := NewAuth(writeTmpfile(testfile))
	assert.NoError(t, err)

	testUsers := []string{"bob-md5", "bob-bcrypt", "bob-sha"}
	for _, name := range testUsers {
		t.Run(name, func(t *testing.T) {
			authenticated, err := auth.Authenticate(name, "secret")
			assert.NoError(t, err)
			assert.True(t, authenticated)

			authenticated, err = auth.Authenticate(name, "XXXXX")
			assert.NoError(t, err)
			assert.False(t, authenticated)
		})
	}
}

func TestClient_UnknownUser(t *testing.T) {
	auth, err := NewAuth(writeTmpfile(testfile))
	assert.NoError(t, err)

	authenticated, err := auth.Authenticate("unknown", "secret")
	assert.NoError(t, err)
	assert.False(t, authenticated)
}

func TestClient_ErrorOnMissingFile(t *testing.T) {
	_, err := NewAuth("/tmp/foo/bar/nothing")
	assert.Error(t, err)
}

func TestClient_ErrorOnInvalidFileContents(t *testing.T) {
	_, err := NewAuth(writeTmpfile("foo bar bazz"))
	assert.Error(t, err)

	_, err = NewAuth(writeTmpfile("foo:bar\nfoo:bar:bazz"))
	assert.Error(t, err)
}

func TestClient_BadMD5Format(t *testing.T) {
	// missing $ separator in md5 hash
	a, err := NewAuth(writeTmpfile("foo:$apr1$IDZSCL/oN68zaFDDRivjour94OVeB."))
	assert.NoError(t, err)

	authenticated, err := a.Authenticate("foo", "secret")
	assert.NoError(t, err)
	assert.False(t, authenticated)
}

func TestClient_Hashes_UnknownAlgoError(t *testing.T) {
	auth, err := NewAuth(writeTmpfile(testfile))
	assert.NoError(t, err)

	authenticated, err := auth.Authenticate("bob-foo", "secret")
	assert.Error(t, err)
	assert.False(t, authenticated)
}

func writeTmpfile(contents string) string {
	f, err := ioutil.TempFile("", "loginsrv_htpasswdtest")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(contents)
	if err != nil {
		panic(err)
	}
	return f.Name()
}
