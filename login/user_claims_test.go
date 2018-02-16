package login

import (
	. "github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/model"
	"io/ioutil"
	"os"
	"testing"
)

var claimsExample = `
- sub: bob
  origin: htpasswd
  claims:
    role: superAdmin

- email: admin@example.org
  origin: google
  claims:
    role: admin
    projects:
      - example
    sub: overwrittenSubject

- domain: example.org
  origin: google
  claims:
    role: user
    projects:
      - example

- claims:
    role: unknown
`

func Test_UserClaims_ParseUserClaims(t *testing.T) {
	f, _ := ioutil.TempFile("", "")
	f.WriteString(claimsExample)
	f.Close()
	defer os.Remove(f.Name())

	c, err := NewUserClaims(&Config{UserFile: f.Name()})
	NoError(t, err)
	Equal(t, 4, len(c.userFileEntries))
	Equal(t, "admin@example.org", c.userFileEntries[1].Email)
	Equal(t, "google", c.userFileEntries[1].Origin)
	Equal(t, "admin", c.userFileEntries[1].Claims["role"])
	Equal(t, []interface{}{"example"}, c.userFileEntries[1].Claims["projects"])
}

func Test_UserClaims_Claims(t *testing.T) {
	f, _ := ioutil.TempFile("", "")
	f.WriteString(claimsExample)
	f.Close()
	defer os.Remove(f.Name())

	c, err := NewUserClaims(&Config{UserFile: f.Name()})
	NoError(t, err)

	// Match first entry
	claims, _ := c.Claims(model.UserInfo{Sub: "bob", Origin: "htpasswd"})
	Equal(t, customClaims{"sub": "bob", "origin": "htpasswd", "role": "superAdmin"}, claims)

	// Match second entry
	claims, _ = c.Claims(model.UserInfo{Sub: "any", Email: "admin@example.org", Origin: "google"})
	Equal(t, customClaims{"sub": "overwrittenSubject", "email": "admin@example.org", "origin": "google", "role": "admin", "projects": []interface{}{"example"}}, claims)

	// default case with no rules
	claims, _ = c.Claims(model.UserInfo{Sub: "bob"})
	Equal(t, customClaims{"sub": "bob", "role": "unknown"}, claims)
}

func Test_UserClaims_NoMatch(t *testing.T) {
	f, _ := ioutil.TempFile("", "")
	f.WriteString(`
- sub: bob
  claims:
    role: superAdmin
`)
	f.Close()
	defer os.Remove(f.Name())

	c, err := NewUserClaims(&Config{UserFile: f.Name()})
	NoError(t, err)

	// Mo Match -> not Modified
	claims, err := c.Claims(model.UserInfo{Sub: "foo"})
	NoError(t, err)
	Equal(t, model.UserInfo{Sub: "foo"}, claims)
}
