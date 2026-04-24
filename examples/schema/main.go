package main

import (
	"fmt"

	gocsv "github.com/mukbeast4/go-csv"
	"github.com/mukbeast4/go-csv/schema"
)

func main() {
	f := gocsv.NewFile()
	f.SetHeaders([]string{"id", "email", "age", "role"})
	f.AppendRow([]any{1, "alice@a.com", 30, "admin"})
	f.AppendRow([]any{2, "invalid", -5, "hacker"})
	f.AppendRow([]any{3, "bob@b.com", 28, "user"})

	s := schema.New().
		Col("id", schema.Required(), schema.Int()).
		Col("email", schema.Required(), schema.Regex(`^[^@]+@[^@]+$`)).
		Col("age", schema.Int(), schema.Range(0, 150)).
		Col("role", schema.OneOf("admin", "user", "guest"))

	errs := s.Validate(f)
	if len(errs) == 0 {
		fmt.Println("All rows valid")
		return
	}
	for _, e := range errs {
		fmt.Printf("row %d field %q [%s]: %v\n", e.Row, e.Field, e.Rule, e.Err)
	}
}
