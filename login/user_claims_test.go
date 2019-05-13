package login

import (
	"testing"
	"time"

	. "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewUserClaims_File(t *testing.T) {
	userFile, cleanup := createClaimsFile(`
- sub: bob
  claims:
    role: superAdmin
`)
	defer cleanup()
	config := &Config{UserFile: userFile}

	claims, err := NewUserClaims(config)

	require.NoError(t, err)
	NotNil(t, claims)
	_, ok := claims.(*userClaimsFile)
	True(t, ok)
}

func Test_NewUserClaims_Provider(t *testing.T) {
	config := &Config{
		UserEndpoint:        "https://test.io/something",
		UserEndpointToken:   "token",
		UserEndpointTimeout: time.Minute,
	}

	claims, err := NewUserClaims(config)

	require.NoError(t, err)
	NotNil(t, claims)
	_, ok := claims.(*userClaimsProvider)
	True(t, ok)
}

func Test_customClaims_Valid(t *testing.T) {
	cc := customClaims{
		"exp": time.Now().Unix() + 3600,
	}

	err := cc.Valid()

	NoError(t, err)
}

func Test_customClaims_Invalid(t *testing.T) {
	cc := customClaims{
		"exp": time.Now().Unix() - 3600,
	}

	err := cc.Valid()

	Error(t, err)
}
