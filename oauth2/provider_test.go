package oauth2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ProviderRegistration(t *testing.T) {
	github, exist := GetProvider("github")
	assert.NotNil(t, github)
	assert.True(t, exist)

	list := ProviderList()
	assert.Equal(t, 1, len(list))
	assert.Equal(t, "github", list[0])
}
