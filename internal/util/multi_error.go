package util

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
