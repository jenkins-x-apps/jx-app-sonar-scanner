package util

import (
	"fmt"
	"strings"
	"testing"

	"github.com/magiconair/properties/assert"
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

func TestIsNotEmpty(t *testing.T) {
	type args struct {
		value interface{}
		key   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"notempty", args{value: "somevalue", key: "somekey"}, false},
		{"empty", args{value: "", key: "somekey"}, true},
		{"nil", args{value: nil, key: "somekey"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := IsNotEmpty(tt.args.value, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("IsNotEmpty() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsInt(t *testing.T) {
	type args struct {
		value interface{}
		key   string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"int", args{value: "3", key: "somekey"}, false},
		{"notint", args{value: "string", key: "somekey"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := IsInt(tt.args.value, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("IsInt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
