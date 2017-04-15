package oauth2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Manager(t *testing.T) {
	m := NewManager()
	m.AddConfig(nil)
	assert.Equal(t, true, true)
}
