package schema

import (
	"fmt"
	"strings"
)

type FieldError struct {
	Row   int
	Field string
	Value string
	Rule  string
	Err   error
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("row %d field %q rule %s: %v", e.Row, e.Field, e.Rule, e.Err)
}

func (e *FieldError) Unwrap() error { return e.Err }

type ValidationError struct {
	Errors []*FieldError
}

func (e *ValidationError) Error() string {
	parts := make([]string, len(e.Errors))
	for i, fe := range e.Errors {
		parts[i] = fe.Error()
	}
	return fmt.Sprintf("validation failed (%d errors): %s", len(e.Errors), strings.Join(parts, "; "))
}

func (e *ValidationError) HasErrors() bool {
	return len(e.Errors) > 0
}
