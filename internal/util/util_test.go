package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/udhos/equalfile"
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

func Test_appropriateToScan(t *testing.T) {
	type args struct {
		infrastructure bool
		pipelineKind   string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"pullrequest", args{false, "pullrequest"}, true},
		{"release", args{false, "release"}, true},
		{"infrastructure", args{true, "release"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := appropriateToScan(tt.args.infrastructure, tt.args.pipelineKind); got != tt.want {
				t.Errorf("appropriateToScan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	testDataLocation := "../../test/"
	testRunName := "run-copyfile"
	original := filepath.Join(testDataLocation, "files", "original")

	// Create a temporary directory for testing and ensure it is cleaned up after
	tmpDirs := make([]string, 0)
	defer func() {
		for _, dir := range tmpDirs {
			err := os.RemoveAll(dir)
			assert.NoError(t, err)
		}
	}()

	dir, err := ioutil.TempDir(testDataLocation, testRunName)
	tmpDirs = append(tmpDirs, dir)
	assert.NoError(t, err)
	copy := filepath.Join(dir, "copy")

	err = CopyFile(original, copy)
	if err != nil {
		t.Errorf("CopyFile() error = %v", err)
		return
	}
	cmp := equalfile.New(nil, equalfile.Options{})
	equal, err := cmp.CompareFile(original, copy)
	if err != nil {
		t.Errorf("CompareFile() error = %v", err)
		return
	}
	assert.True(t, equal)
}

func TestFileExists(t *testing.T) {
	assert.True(t, FileExists("../../test/files/original"))
	assert.False(t, FileExists("../../test/files/pinkelephant"))
}
