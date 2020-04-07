package util

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/cenkalti/backoff"
)

var timeout = 60 * time.Second

// Contains checks whether the specified string is contained in the given string slice.
// Returns true if it does, false otherwise
func Contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// ApplyWithBackoff tries to apply the specified function using an exponential backoff algorithm.
// If the function eventually succeed nil is returned, otherwise the error returned by f.
func ApplyWithBackoff(f func() error) error {
	exponentialBackOff := backoff.NewExponentialBackOff()
	exponentialBackOff.MaxElapsedTime = timeout
	exponentialBackOff.Reset()
	return backoff.Retry(f, exponentialBackOff)
}

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func CopyFile(src string, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

// MoveFile renames the file named src to the file named by dst.
// If dst already exists and is not a directory, MoveFile replaces it
func MoveFile(src string, dst string) (err error) {
	return os.Rename(src, dst)
}

// FileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// AppropriateToScan checks the current pipeline execution environment to see
// if it is appropriate to insert the app.
func AppropriateToScan() bool {
	configFileExists := FileExists(".pre-commit-config.yaml")
	pipelineKind := os.Getenv("PIPELINE_KIND")
	return appropriateToScan(configFileExists, pipelineKind)
}

func appropriateToScan(infrastructure bool, pipelineKind string) bool {
	// Should we be attempting to patch in this execution scope?
	if pipelineKind == "pullrequest" && !infrastructure {
		fmt.Println("Detected preview build. Preparing to scan...")
		return true
	} else if pipelineKind == "release" && !infrastructure {
		fmt.Println("Detected release build. Preparing to scan...")
		return true
	} else {
		// Environment build so skip
		fmt.Println("Skipping sonar-scan")
		return false
	}
}
