package oauth2

import (
	. "github.com/stretchr/testify/assert"
	"testing"
)

func Test_ProviderRegistration(t *testing.T) {
	github, exist := GetProvider("github")
	NotNil(t, github)
	True(t, exist)

	google, exist := GetProvider("google")
	NotNil(t, google)
	True(t, exist)

	list := ProviderList()
	Equal(t, 2, len(list))
	Contains(t, list, "github")
	Contains(t, list, "google")
}
