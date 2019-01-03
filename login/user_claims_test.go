package login

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	. "github.com/stretchr/testify/assert"
	"github.com/tarent/loginsrv/model"
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

- origin: gitlab
  groups:
    - "example/subgroup"
    - othergroup
  claims:
    role: admin

- claims:
    role: unknown
`

var invalidClaimsExample = `
- sub: bob
	origin: google
`

func Test_ParseUserClaims_InvalidFile(t *testing.T) {
	c, err := NewUserClaims(&Config{UserFile: "notfound"})
	Error(t, err)
	Equal(t, &UserClaims{
		userFile:        "notfound",
		userFileEntries: []userFileEntry{},
	}, c)
}

func Test_ParseUserClaims_InvalidYAML(t *testing.T) {
	f, _ := ioutil.TempFile("", "")
	f.WriteString(invalidClaimsExample)
	f.Close()
	defer os.Remove(f.Name())

	c, err := NewUserClaims(&Config{UserFile: f.Name()})
	Error(t, err)
	Equal(t, &UserClaims{
		userFile:        f.Name(),
		userFileEntries: []userFileEntry{},
	}, c)
}

func Test_UserClaims_ParseUserClaims(t *testing.T) {
	f, _ := ioutil.TempFile("", "")
	f.WriteString(claimsExample)
	f.Close()
	defer os.Remove(f.Name())

	c, err := NewUserClaims(&Config{UserFile: f.Name()})
	NoError(t, err)
	Equal(t, 5, len(c.userFileEntries))
	Equal(t, "admin@example.org", c.userFileEntries[1].Email)
	Equal(t, "google", c.userFileEntries[1].Origin)
	Equal(t, "admin", c.userFileEntries[1].Claims["role"])
	Equal(t, []interface{}{"example"}, c.userFileEntries[1].Claims["projects"])
	Equal(t, []string{"example/subgroup", "othergroup"}, c.userFileEntries[3].Groups)
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

	// Match fourth entry
	claims, _ = c.Claims(model.UserInfo{Sub: "any", Groups: []string{"example/subgroup", "othergroup"}, Origin: "gitlab"})
	Equal(t, customClaims{"sub": "any", "groups": []string{"example/subgroup", "othergroup"}, "origin": "gitlab", "role": "admin"}, claims)

	// default case with no rules
	claims, _ = c.Claims(model.UserInfo{Sub: "bob"})
	Equal(t, customClaims{"sub": "bob", "role": "unknown"}, claims)
}

func Test_UserClaims_NoMatch(t *testing.T) {
	f, _ := ioutil.TempFile("", "")
	f.WriteString(`
- sub: bob
  groups:
    - othergroup
  claims:
    role: superAdmin
`)
	f.Close()
	defer os.Remove(f.Name())

	c, err := NewUserClaims(&Config{UserFile: f.Name()})
	NoError(t, err)

	// No Match -> not Modified
	claims, err := c.Claims(model.UserInfo{Sub: "foo"})
	NoError(t, err)
	Equal(t, model.UserInfo{Sub: "foo"}, claims)

	claims, err = c.Claims(model.UserInfo{Sub: "bob", Groups: []string{"group"}})
	NoError(t, err)
	Equal(t, model.UserInfo{Sub: "bob", Groups: []string{"group"}}, claims)
}

func Test_UserClaims_Valid(t *testing.T) {
	cc := customClaims{
		"exp": time.Now().Unix() + 3600,
	}

	err := cc.Valid()
	NoError(t, err)
}

func Test_UserClaims_Invalid(t *testing.T) {
	cc := customClaims{
		"exp": time.Now().Unix() - 3600,
	}

	err := cc.Valid()
	Error(t, err)
}
