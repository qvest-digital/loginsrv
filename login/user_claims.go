package login

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/tarent/loginsrv/model"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

type customClaims map[string]interface{}

func (custom customClaims) Valid() error {
	if exp, ok := custom["exp"]; ok {
		if exp, ok := exp.(int64); ok {
			if exp < time.Now().Unix() {
				return errors.New("token expired")
			}
		}
	}
	return nil
}

type userFileEntry struct {
	Sub    string                 `yaml:"sub"`
	Origin string                 `yaml:"origin"`
	Email  string                 `yaml:"email"`
	Domain string                 `yaml:"domain"`
	Claims map[string]interface{} `yaml:"claims"`
}

type UserClaims struct {
	userFile        string
	userFileEntries []userFileEntry
}

func NewUserClaims(config *Config) (*UserClaims, error) {
	c := &UserClaims{
		userFile:        config.UserFile,
		userFileEntries: []userFileEntry{},
	}
	err := c.parseUserFile()
	return c, err
}

func (c *UserClaims) parseUserFile() error {
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
func (c *UserClaims) Claims(userInfo model.UserInfo) (jwt.Claims, error) {
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
	return true
}
