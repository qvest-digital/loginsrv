package login

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_ParseBackendOptions(t *testing.T) {
	testCases := []struct {
		input       []string
		expected    BackendOptions
		expectError bool
	}{
		{
			[]string{},
			BackendOptions{},
			false,
		},
		{
			[]string{"name=p1,key1=value1,key2=value2"},
			BackendOptions{},
			true, // no provider name specified
		},
		{
			[]string{
				"provider=simple,name=p1,key1=value1,key2=value2",
				"provider=simple,name=p2,key3=value3,key4=value4",
			},
			BackendOptions{
				map[string]string{
					"provider": "simple",
					"name":     "p1",
					"key1":     "value1",
					"key2":     "value2",
				},
				map[string]string{
					"provider": "simple",
					"name":     "p2",
					"key3":     "value3",
					"key4":     "value4",
				},
			},
			false,
		},
		{
			[]string{"foo"},
			BackendOptions{},
			true,
		},
	}
	for i, test := range testCases {
		t.Run(fmt.Sprintf("test %v", i), func(t *testing.T) {
			options := &BackendOptions{}
			for _, input := range test.input {
				err := options.Set(input)
				if test.expectError {
					assert.Error(t, err)
				} else {
					if err != nil {
						assert.NoError(t, err)
						continue
					}
				}
			}
			assert.Equal(t, test.expected, *options)
		})
	}
}
