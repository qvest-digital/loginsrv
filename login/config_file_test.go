package login

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigFileSimple(t *testing.T) {
	config := &Config{
		ConfigFile: "testdata/test1.yaml",
	}
	err := parseConfigFile(config)
	assert.NoError(t, err)
}
