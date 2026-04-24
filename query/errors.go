package query

import "fmt"

type MissingColumnError struct {
	Column string
}

func (e *MissingColumnError) Error() string {
	return fmt.Sprintf("query: column %q not found", e.Column)
}
