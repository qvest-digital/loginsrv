package login

import (
	"io/ioutil"
	"os"
	"testing"

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

func Test_newUserClaimsFile_InvalidFile(t *testing.T) {
	c, err := newUserClaimsFile("notfound")

	Error(t, err)
	Equal(t, &userClaimsFile{
		userFile:        "notfound",
		userFileEntries: []userFileEntry{},
	}, c)
}

func Test_newUserClaimsFile_InvalidYAML(t *testing.T) {
	f, _ := ioutil.TempFile("", "")
	f.WriteString(invalidClaimsExample)
	f.Close()
	defer os.Remove(f.Name())

	c, err := newUserClaimsFile(f.Name())

	Error(t, err)
	Equal(t, &userClaimsFile{
		userFile:        f.Name(),
		userFileEntries: []userFileEntry{},
	}, c)
}

func Test_newUserClaimsFile_ParseFile(t *testing.T) {
	fileName, cleanup := createClaimsFile(claimsExample)
	defer cleanup()

	c, err := newUserClaimsFile(fileName)

	NoError(t, err)
	Equal(t, 5, len(c.userFileEntries))
	Equal(t, "admin@example.org", c.userFileEntries[1].Email)
	Equal(t, "google", c.userFileEntries[1].Origin)
	Equal(t, "admin", c.userFileEntries[1].Claims["role"])
	Equal(t, []interface{}{"example"}, c.userFileEntries[1].Claims["projects"])
	Equal(t, []string{"example/subgroup", "othergroup"}, c.userFileEntries[3].Groups)
}

func Test_userClaimsFile_Claims(t *testing.T) {
	f, _ := ioutil.TempFile("", "")
	f.WriteString(claimsExample)
	f.Close()
	fileName := f.Name()
	defer os.Remove(f.Name())

	c, err := NewUserClaims(&Config{UserFile: fileName})
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

func Test_userClaimsFile_NoMatch(t *testing.T) {
	userFile, cleanup := createClaimsFile(`
- sub: bob
  groups:
    - othergroup
  claims:
    role: superAdmin
`)
	defer cleanup()

	c, err := NewUserClaims(&Config{UserFile: userFile})
	NoError(t, err)

	// No Match -> not Modified
	claims, err := c.Claims(model.UserInfo{Sub: "foo"})
	NoError(t, err)
	Equal(t, model.UserInfo{Sub: "foo"}, claims)

	claims, err = c.Claims(model.UserInfo{Sub: "bob", Groups: []string{"group"}})
	NoError(t, err)
	Equal(t, model.UserInfo{Sub: "bob", Groups: []string{"group"}}, claims)
}

func createClaimsFile(claims string) (string, func()) {
	f, _ := ioutil.TempFile("", "")
	f.WriteString(claims)
	f.Close()

	return f.Name(), func() { os.Remove(f.Name()) }
}
