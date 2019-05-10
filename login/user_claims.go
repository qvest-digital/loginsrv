package login

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/tarent/loginsrv/model"
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

func (custom customClaims) merge(values map[string]interface{}) {
	for k, v := range values {
		custom[k] = v
	}
}

type UserClaims interface {
	Claims(userInfo model.UserInfo) (jwt.Claims, error)
}

func NewUserClaims(config *Config) (UserClaims, error) {
	if config.UserEndpoint != "" {
		return newUserClaimsProvider(config.UserEndpoint, config.UserEndpointToken, config.UserEndpointTimeout)
	}
	return newUserClaimsFile(config.UserFile)
}
