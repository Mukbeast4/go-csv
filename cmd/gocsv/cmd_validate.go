package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	gocsv "github.com/mukbeast4/go-csv"
	"github.com/mukbeast4/go-csv/schema"
)

type schemaFile struct {
	Columns []columnSpec `json:"columns"`
}

type columnSpec struct {
	Name     string   `json:"name"`
	Type     string   `json:"type,omitempty"`
	Required bool     `json:"required,omitempty"`
	Regex    string   `json:"regex,omitempty"`
	Min      *float64 `json:"min,omitempty"`
	Max      *float64 `json:"max,omitempty"`
	OneOf    []string `json:"one_of,omitempty"`
}

func cmdValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	schemaPath := fs.String("schema", "", "path to JSON schema file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *schemaPath == "" || fs.NArg() < 1 {
		return fmt.Errorf("usage: validate -schema schema.json file.csv")
	}
	data, err := os.ReadFile(*schemaPath)
	if err != nil {
		return err
	}
	var sf schemaFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return err
	}
	s := buildSchema(sf)
	f, err := gocsv.OpenFile(fs.Arg(0), gocsv.WithHeader(true))
	if err != nil {
		return err
	}
	errs := s.Validate(f)
	if len(errs) == 0 {
		fmt.Println("OK")
		return nil
	}
	for _, e := range errs {
		fmt.Printf("row %d field %q [%s]: %v\n", e.Row, e.Field, e.Rule, e.Err)
	}
	return fmt.Errorf("%d validation errors", len(errs))
}

func buildSchema(sf schemaFile) *schema.Schema {
	s := schema.New()
	for _, c := range sf.Columns {
		var rules []schema.Rule
		if c.Required {
			rules = append(rules, schema.Required())
		}
		switch c.Type {
		case "int":
			rules = append(rules, schema.Int())
		case "float":
			rules = append(rules, schema.Float())
		case "bool":
			rules = append(rules, schema.Bool())
		case "string":
			rules = append(rules, schema.String())
		}
		if c.Regex != "" {
			rules = append(rules, schema.Regex(c.Regex))
		}
		if c.Min != nil && c.Max != nil {
			rules = append(rules, schema.Range(*c.Min, *c.Max))
		} else if c.Min != nil {
			rules = append(rules, schema.Min(*c.Min))
		} else if c.Max != nil {
			rules = append(rules, schema.Max(*c.Max))
		}
		if len(c.OneOf) > 0 {
			rules = append(rules, schema.OneOf(c.OneOf...))
		}
		s.Col(c.Name, rules...)
	}
	return s
}
