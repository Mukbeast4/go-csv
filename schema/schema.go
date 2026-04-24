package schema

import (
	gocsv "github.com/mukbeast4/go-csv"
)

type column struct {
	Name  string
	Rules []Rule
}

type Schema struct {
	cols []column
}

func New() *Schema {
	return &Schema{}
}

func (s *Schema) Col(name string, rules ...Rule) *Schema {
	s.cols = append(s.cols, column{Name: name, Rules: rules})
	return s
}

func (s *Schema) Columns() []string {
	out := make([]string, len(s.cols))
	for i, c := range s.cols {
		out[i] = c.Name
	}
	return out
}

func (s *Schema) Validate(f *gocsv.File) []*FieldError {
	var errs []*FieldError
	headers := f.Headers()
	colIdx := buildIdx(headers)
	rows, _ := f.GetRows()
	for rowI, row := range rows {
		for _, c := range s.cols {
			idx, ok := colIdx[c.Name]
			value := ""
			if ok && idx < len(row) {
				value = row[idx]
			}
			if !ok {
				value = ""
			}
			for _, rule := range c.Rules {
				if err := rule.Check(value); err != nil {
					errs = append(errs, &FieldError{
						Row:   rowI,
						Field: c.Name,
						Value: value,
						Rule:  rule.Name(),
						Err:   err,
					})
				}
			}
		}
	}
	return errs
}

func (s *Schema) ValidateOrErr(f *gocsv.File) error {
	errs := s.Validate(f)
	if len(errs) == 0 {
		return nil
	}
	return &ValidationError{Errors: errs}
}

func (s *Schema) ValidateStream(it *gocsv.RowIterator) <-chan *FieldError {
	out := make(chan *FieldError, 64)
	headers := it.Headers()
	colIdx := buildIdx(headers)
	go func() {
		defer close(out)
		rowI := 0
		for it.Next() {
			row := it.Row()
			for _, c := range s.cols {
				idx, ok := colIdx[c.Name]
				value := ""
				if ok && idx < len(row) {
					value = row[idx]
				}
				for _, rule := range c.Rules {
					if err := rule.Check(value); err != nil {
						out <- &FieldError{
							Row:   rowI,
							Field: c.Name,
							Value: value,
							Rule:  rule.Name(),
							Err:   err,
						}
					}
				}
			}
			rowI++
		}
	}()
	return out
}

func buildIdx(headers []string) map[string]int {
	idx := make(map[string]int, len(headers))
	for i, h := range headers {
		idx[h] = i
	}
	return idx
}
