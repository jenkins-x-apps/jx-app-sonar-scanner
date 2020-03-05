package util

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestContains(t *testing.T) {
	var testCases = []struct {
		slice          []string
		element        string
		expectedResult bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "e", false},
		{nil, "b", false},
	}

	for _, testCase := range testCases {
		actualResult := Contains(testCase.slice, testCase.element)
		assert.Equal(t, testCase.expectedResult, actualResult)
	}
}

func TestApplyWithBackoffFailure(t *testing.T) {
	origTimeout := timeout
	defer func() {
		timeout = origTimeout
	}()
	timeout = 1 * time.Second

	var callCount = 0
	f := func() error {
		callCount++
		return errors.New("bang")
	}
	err := ApplyWithBackoff(f)

	assert.Error(t, err)
	assert.True(t, callCount > 1)
}

func TestApplyWithBackoffSuccess(t *testing.T) {
	origTimeout := timeout
	defer func() {
		timeout = origTimeout
	}()
	timeout = 10 * time.Second

	var callCount = 0
	f := func() error {
		if callCount == 3 {
			return nil
		}
		callCount++
		return errors.New("bang")
	}
	err := ApplyWithBackoff(f)

	assert.NoError(t, err)
	assert.Equal(t, 3, callCount)
}
