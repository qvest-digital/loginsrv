package model

import (
	. "github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_UserInfo_Valid(t *testing.T) {
	Error(t, UserInfo{Expiry: 0}.Valid())
	Error(t, UserInfo{Expiry: time.Now().Add(-1 * time.Second).Unix()}.Valid())
	NoError(t, UserInfo{Expiry: time.Now().Add(time.Second).Unix()}.Valid())
}
