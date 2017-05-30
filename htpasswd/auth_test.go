package htpasswd

import (
	. "github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
	"time"
)

// password for all of them is 'secret'
const testfile = `bob-md5:$apr1$IDZSCL/o$N68zaFDDRivjour94OVeB.

bob-bcrypt:$2y$05$Hw6y1sFwh6CdwiPOKFMYj..xVSQWI3wzyQvt5th392ig8RLmeLU.6
bob-sha:{SHA}5en6G6MezRroT3XKqkdPOmY/BfQ=

# a comment
bob-foo:{fooo}sdcsdcsdc/BfQ=


`

func TestAuth_Hashes(t *testing.T) {
	auth, err := NewAuth(writeTmpfile(testfile))
	NoError(t, err)

	testUsers := []string{"bob-md5", "bob-bcrypt", "bob-sha"}
	for _, name := range testUsers {
		t.Run(name, func(t *testing.T) {
			authenticated, err := auth.Authenticate(name, "secret")
			NoError(t, err)
			True(t, authenticated)

			authenticated, err = auth.Authenticate(name, "XXXXX")
			NoError(t, err)
			False(t, authenticated)
		})
	}
}

func TestAuth_ReloadFile(t *testing.T) {
	filename := writeTmpfile(`bob:$apr1$IDZSCL/o$N68zaFDDRivjour94OVeB.`)
	auth, err := NewAuth(filename)
	NoError(t, err)

	authenticated, err := auth.Authenticate("bob", "secret")
	NoError(t, err)
	True(t, authenticated)

	authenticated, err = auth.Authenticate("alice", "secret")
	NoError(t, err)
	False(t, authenticated)

	// The refresh is time based, so we have to wait a second, here
	time.Sleep(time.Second)

	err = ioutil.WriteFile(filename, []byte(`alice:$apr1$IDZSCL/o$N68zaFDDRivjour94OVeB.`), 06644)
	NoError(t, err)

	authenticated, err = auth.Authenticate("bob", "secret")
	NoError(t, err)
	False(t, authenticated)

	authenticated, err = auth.Authenticate("alice", "secret")
	NoError(t, err)
	True(t, authenticated)
}

func TestAuth_UnknownUser(t *testing.T) {
	auth, err := NewAuth(writeTmpfile(testfile))
	NoError(t, err)

	authenticated, err := auth.Authenticate("unknown", "secret")
	NoError(t, err)
	False(t, authenticated)
}

func TestAuth_ErrorOnMissingFile(t *testing.T) {
	_, err := NewAuth("/tmp/foo/bar/nothing")
	Error(t, err)
}

func TestAuth_ErrorOnInvalidFileContents(t *testing.T) {
	_, err := NewAuth(writeTmpfile("foo bar bazz"))
	Error(t, err)

	_, err = NewAuth(writeTmpfile("foo:bar\nfoo:bar:bazz"))
	Error(t, err)
}

func TestAuth_BadMD5Format(t *testing.T) {
	// missing $ separator in md5 hash
	a, err := NewAuth(writeTmpfile("foo:$apr1$IDZSCL/oN68zaFDDRivjour94OVeB."))
	NoError(t, err)

	authenticated, err := a.Authenticate("foo", "secret")
	NoError(t, err)
	False(t, authenticated)
}

func TestAuth_Hashes_UnknownAlgoError(t *testing.T) {
	auth, err := NewAuth(writeTmpfile(testfile))
	NoError(t, err)

	authenticated, err := auth.Authenticate("bob-foo", "secret")
	Error(t, err)
	False(t, authenticated)
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
