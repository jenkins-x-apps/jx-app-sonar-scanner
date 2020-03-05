package util

import (
	"fmt"
	"strings"
)

// MultiError is a collection of errors.
type MultiError struct {
	Errors []error
}

// Empty returns true if current MuiltiError is empty,
// false otherwise.
func (m *MultiError) Empty() bool {
	return len(m.Errors) == 0
}

// Collect appends an error to this MultiError.
func (m *MultiError) Collect(err error) {
	if err != nil {
		m.Errors = append(m.Errors, err)
	}
}

// ToError returns a single error made up of all errors in this MultiError.
func (m MultiError) ToError() error {
	if len(m.Errors) == 0 {
		return nil
	}

	errStrings := []string{}
	for _, err := range m.Errors {
		errStrings = append(errStrings, err.Error())
	}
	return fmt.Errorf(strings.Join(errStrings, "\n"))
}
