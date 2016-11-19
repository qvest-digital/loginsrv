package osiam

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var testToken = &Token{
	Userid:                "84f6cffa-4505-48ec-a851-424160892283",
	ExpiresIn:             28493,
	ExpiresAt:             Timestamp{time.Unix(1479600891424, 0)},
	RefreshToken:          "15b22304-f838-48c2-9c40-18bf285060a6",
	ClientId:              "example-client",
	RefreshTokenExpiresAt: Timestamp{time.Unix(1479575304893, 0)},
	UserName:              "admin",
	TokenType:             "bearer",
	AccessToken:           "59f39ef8-1dc3-4c0d-8dea-c9597ef0a8ef",
	Scope:                 "ME",
}

var testTokenString = `{
                "user_id" : "84f6cffa-4505-48ec-a851-424160892283",
                "expires_in" : 28493,
                "expires_at" : 1479600891424,
                "refresh_token" : "15b22304-f838-48c2-9c40-18bf285060a6",
                "client_id" : "example-client",
                "refresh_token_expires_at" : 1479575304893,
                "user_name" : "admin",
                "token_type" : "bearer",
                "access_token" : "59f39ef8-1dc3-4c0d-8dea-c9597ef0a8ef",
                "scope" : "ME"}`

func TestClient_TokenMarshaling(t *testing.T) {
	tk := &Token{}
	json.Unmarshal([]byte(testTokenString), tk)
	assert.Equal(t, testToken, tk)

	_, err := json.Marshal(tk)
	assert.NoError(t, err)
}
