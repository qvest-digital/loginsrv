package osiam

import (
	"testing"

	. "github.com/stretchr/testify/assert"
)

func TestError_ParseOsiamError_CornerCases(t *testing.T) {
	e := ParseOsiamError([]byte(""))
	Equal(t, "client_parse_error", e.Error)

	e = ParseOsiamError([]byte("{}"))
	Equal(t, "client_parse_error", e.Error)
}
