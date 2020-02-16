package saml

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

func Test_SamlProviderRegistration(t *testing.T) {
	azure, exist := GetProvider("azure")
	NotNil(t, azure)
	True(t, exist)

	list := ProviderList()
	Equal(t, 1, len(list))
	Contains(t, list, "azure")
}
