package oauth2

import (
	. "github.com/stretchr/testify/assert"
	"testing"
)

func Test_ProviderRegistration(t *testing.T) {
	github, exist := GetProvider("github")
	NotNil(t, github)
	True(t, exist)

	list := ProviderList()
	Equal(t, 1, len(list))
	Equal(t, "github", list[0])
}
