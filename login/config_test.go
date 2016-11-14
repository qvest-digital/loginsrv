package login

import (
//"fmt"
//"github.com/stretchr/testify/assert"
//"testing"
)

/**
func TestConfig_GetBackendOptions(t *testing.T) {
	testCases := []struct {
		backends      []string
		parsedOptions BackendOptions
		expectError   bool
	}{
		{
			[]string{},
			[]map[string]string{},
			false,
		},
		{
			[]string{
				"name=p1,key1=value1,key2=value2",
				"name=p2,key3=value3,key4=value4",
			},
			[]map[string]string{
				map[string]string{
					"name": "p1",
					"key1": "value1",
					"key2": "value2",
				},
				map[string]string{
					"name": "p2",
					"key3": "value3",
					"key4": "value4",
				},
			},
			false,
		},
		{
			[]string{"foo"},
			nil,
			true,
		},
	}
	for i, test := range testCases {
		t.Run(fmt.Sprintf("test %v", i), func(t *testing.T) {
			cfg := &Config{}
			cfg.Backends = test.backends
			opts, err := cfg.GetBackendOptions()
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, opts, test.parsedOptions)
			}
		})
	}
}
**/
