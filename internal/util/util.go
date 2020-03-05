package util

import (
	"github.com/cenkalti/backoff"
	"io"
	"os"
	"time"
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
