package login

import (
	"io/ioutil"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/tarent/loginsrv/model"
	yaml "gopkg.in/yaml.v2"
)

type userFileEntry struct {
	Sub    string                 `yaml:"sub"`
	Origin string                 `yaml:"origin"`
	Email  string                 `yaml:"email"`
	Domain string                 `yaml:"domain"`
	Groups []string               `yaml:"groups"`
	Claims map[string]interface{} `yaml:"claims"`
}

type userClaimsFile struct {
	userFile        string
	userFileEntries []userFileEntry
}

func newUserClaimsFile(file string) (*userClaimsFile, error) {
	c := &userClaimsFile{
		userFile:        file,
		userFileEntries: []userFileEntry{},
	}
	err := c.parseUserFile()
	return c, err
}

func (c *userClaimsFile) parseUserFile() error {
	if c.userFile == "" {
		return nil
	}
	b, err := ioutil.ReadFile(c.userFile)
	if err != nil {
		return errors.Wrapf(err, "can't read user file %v", c.userFile)
	}

	err = yaml.Unmarshal(b, &c.userFileEntries)
	if err != nil {
		return errors.Wrapf(err, "can't parse user file %v", c.userFile)
	}
	return nil
}

// Claims returns a map of the token claims for a user.
func (c *userClaimsFile) Claims(userInfo model.UserInfo) (jwt.Claims, error) {
	for _, entry := range c.userFileEntries {
		if match(userInfo, entry) {
			claims := customClaims(userInfo.AsMap())
			for k, v := range entry.Claims {
				claims[k] = v
			}
			return claims, nil
		}
	}
	return userInfo, nil
}

func match(userInfo model.UserInfo, entry userFileEntry) bool {
	if entry.Sub != "" && entry.Sub != userInfo.Sub {
		return false
	}
	if entry.Domain != "" && entry.Domain != userInfo.Domain {
		return false
	}
	if entry.Email != "" && entry.Email != userInfo.Email {
		return false
	}
	if entry.Origin != "" && entry.Origin != userInfo.Origin {
		return false
	}
	if len(entry.Groups) > 0 {
		eligible := false
		for _, entryGroup := range entry.Groups {
			for _, userGroup := range userInfo.Groups {
				if entryGroup == userGroup {
					eligible = true
					break
				}
			}
		}
		if !eligible {
			return false
		}
	}
	return true
}
