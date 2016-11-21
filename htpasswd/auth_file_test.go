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
bob-foo:{fooo}sdcsdcsdc/BfQ=
`

func TestClient_Hashes(t *testing.T) {
	auth, err := NewAuth(writeTestfile())
	assert.NoError(t, err)

	//testUsers := []string{"bob-md5", "bob-bcrypt", "bob-sha"}
	testUsers := []string{"bob-bcrypt"}
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

func TestClient_Hashes_UnknownAlgoError(t *testing.T) {
	auth, err := NewAuth(writeTestfile())
	assert.NoError(t, err)

	authenticated, err := auth.Authenticate("bob-foo", "secret")
	assert.Error(t, err)
	assert.False(t, authenticated)
}

func writeTestfile() string {
	f, err := ioutil.TempFile("", "loginsrv_htpasswdtest")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.WriteString(testfile)
	if err != nil {
		panic(err)
	}
	return f.Name()
}
