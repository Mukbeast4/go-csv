package schema

import (
	"testing"

	gocsv "github.com/mukbeast4/go-csv"
)

func fixture() *gocsv.File {
	f := gocsv.NewFile()
	f.SetHeaders([]string{"id", "email", "age", "role"})
	f.AppendRow([]any{1, "alice@a.com", 30, "admin"})
	f.AppendRow([]any{2, "invalid-email", 25, "user"})
	f.AppendRow([]any{3, "bob@b.com", -5, "admin"})
	f.AppendRow([]any{4, "charlie@c.com", 200, "hacker"})
	f.AppendRow([]any{"", "", 28, "user"})
	return f
}

func TestSchemaRequired(t *testing.T) {
	s := New().Col("id", Required())
	errs := s.Validate(fixture())
	if len(errs) != 1 {
		t.Errorf("errs: %d", len(errs))
	}
}

func TestSchemaType(t *testing.T) {
	s := New().Col("age", Int())
	errs := s.Validate(fixture())
	if len(errs) != 0 {
		t.Errorf("should pass: %v", errs)
	}
}

func TestSchemaRange(t *testing.T) {
	s := New().Col("age", Range(0, 150))
	errs := s.Validate(fixture())
	if len(errs) != 2 {
		t.Errorf("expected 2 range errs, got %d: %v", len(errs), errs)
	}
}

func TestSchemaRegex(t *testing.T) {
	s := New().Col("email", Regex(`^[^@]+@[^@]+$`))
	errs := s.Validate(fixture())
	if len(errs) != 1 {
		t.Errorf("expected 1 email err, got %d", len(errs))
	}
}

func TestSchemaOneOf(t *testing.T) {
	s := New().Col("role", OneOf("admin", "user", "guest"))
	errs := s.Validate(fixture())
	if len(errs) != 1 {
		t.Errorf("expected 1 role err, got %d", len(errs))
	}
}

func TestSchemaCombined(t *testing.T) {
	s := New().
		Col("id", Required(), Int()).
		Col("email", Required(), Regex(`^[^@]+@[^@]+$`)).
		Col("age", Int(), Range(0, 150)).
		Col("role", OneOf("admin", "user"))
	errs := s.Validate(fixture())
	if len(errs) < 4 {
		t.Errorf("expected multiple errs, got %d: %v", len(errs), errs)
	}
}

func TestSchemaValidateOrErr(t *testing.T) {
	s := New().Col("age", Range(0, 150))
	err := s.ValidateOrErr(fixture())
	if err == nil {
		t.Error("expected validation error")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if !ve.HasErrors() {
		t.Error("HasErrors should be true")
	}
}

func TestSchemaValidateOK(t *testing.T) {
	f := gocsv.NewFile()
	f.SetHeaders([]string{"id", "name"})
	f.AppendRow([]any{1, "Alice"})
	s := New().Col("id", Int()).Col("name", Required())
	if errs := s.Validate(f); len(errs) != 0 {
		t.Errorf("should pass: %v", errs)
	}
}

func TestSchemaCustom(t *testing.T) {
	s := New().Col("id", Custom("positive", func(v string) error {
		if v == "0" {
			return &customErr{}
		}
		return nil
	}))
	f := gocsv.NewFile()
	f.SetHeaders([]string{"id"})
	f.AppendRow([]any{0})
	errs := s.Validate(f)
	if len(errs) != 1 {
		t.Errorf("expected 1 custom err, got %d", len(errs))
	}
}

type customErr struct{}

func (customErr) Error() string { return "not positive" }

func TestSchemaLength(t *testing.T) {
	s := New().Col("name", MinLen(3), MaxLen(10))
	f := gocsv.NewFile()
	f.SetHeaders([]string{"name"})
	f.AppendRow([]any{"ab"})
	f.AppendRow([]any{"charlie"})
	f.AppendRow([]any{"waaaaaaaaaaaaay too long"})
	errs := s.Validate(f)
	if len(errs) != 2 {
		t.Errorf("got %d: %v", len(errs), errs)
	}
}
