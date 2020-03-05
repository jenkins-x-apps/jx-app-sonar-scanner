package util

import (
	"fmt"
	"github.com/magiconair/properties/assert"
	"strings"
	"testing"
)

func Test_IsBool(t *testing.T) {
	var testBools = []struct {
		value    string
		expected bool
		errors   []string
	}{
		{"true", true, []string{}},
		{"false", false, []string{}},
		{"0", false, []string{}},
		{"1", true, []string{}},
		{"snafu", false, []string{"value for 'FOO' needs to be an bool"}},
		{"", false, []string{"value for 'FOO' needs to be an bool"}},
	}

	for _, testBool := range testBools {
		err := IsBool(testBool.value, "FOO")
		var errors []string
		if err == nil {
			errors = []string{}
		} else {
			errors = strings.Split(err.Error(), "\n")
		}

		assert.Equal(t, testBool.errors, errors, fmt.Sprintf("Unexpected error for %s", testBool.value))
	}
}
