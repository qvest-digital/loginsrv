package model

import (
	"encoding/json"
	. "github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_UserInfo_Valid(t *testing.T) {
	Error(t, UserInfo{Expiry: 0}.Valid())
	Error(t, UserInfo{Expiry: time.Now().Add(-1 * time.Second).Unix()}.Valid())
	NoError(t, UserInfo{Expiry: time.Now().Add(time.Second).Unix()}.Valid())
}

func Test_UserInfo_AsMap(t *testing.T) {
	u := UserInfo{
		Sub:       `json:"sub"`,
		Picture:   `json:"picture,omitempty"`,
		Name:      `json:"name,omitempty"`,
		Email:     `json:"email,omitempty"`,
		Origin:    `json:"origin,omitempty"`,
		Expiry:    23,
		Refreshes: 42,
		Domain:    `json:"domain,omitempty"`,
	}

	givenJson, _ := json.Marshal(u.AsMap())
	given := UserInfo{}
	err := json.Unmarshal(givenJson, &given)
	NoError(t, err)
	Equal(t, u, given)
}

func Test_UserInfo_AsMap_Minimal(t *testing.T) {
	u := UserInfo{
		Sub: `json:"sub"`,
	}

	givenJson, _ := json.Marshal(u.AsMap())
	given := UserInfo{}
	err := json.Unmarshal(givenJson, &given)
	NoError(t, err)
	Equal(t, u, given)
}
