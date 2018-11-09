package oauth2

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

func Test_ProviderRegistration(t *testing.T) {
	github, exist := GetProvider("github")
	NotNil(t, github)
	True(t, exist)

	google, exist := GetProvider("google")
	NotNil(t, google)
	True(t, exist)

	bitbucket, exist := GetProvider("bitbucket")
	NotNil(t, bitbucket)
	True(t, exist)

	facebook, exist := GetProvider("facebook")
	NotNil(t, facebook)
	True(t, exist)

	gitlab, exist := GetProvider("gitlab")
	NotNil(t, gitlab)
	True(t, exist)

	list := ProviderList()
	Equal(t, 5, len(list))
	Contains(t, list, "github")
	Contains(t, list, "google")
	Contains(t, list, "bitbucket")
	Contains(t, list, "facebook")
	Contains(t, list, "gitlab")
}
