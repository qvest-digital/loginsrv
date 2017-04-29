package login

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_ucfirst(t *testing.T) {
	assert.Equal(t, "", ucfirst(""))
	assert.Equal(t, "A", ucfirst("a"))
	assert.Equal(t, "Abc def", ucfirst("abc def"))
}
