package osiam

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestError_ParseOsiamError_CornerCases(t *testing.T) {
	e := ParseOsiamError([]byte(""))
	assert.Equal(t, "client_parse_error", e.Error)

	e = ParseOsiamError([]byte("{}"))
	assert.Equal(t, "client_parse_error", e.Error)
}
