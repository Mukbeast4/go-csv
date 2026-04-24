package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	gocsv "github.com/mukbeast4/go-csv"
)

func cmdGenStruct(args []string) error {
	fs := flag.NewFlagSet("gen-struct", flag.ContinueOnError)
	name := fs.String("name", "Row", "struct name")
	pkg := fs.String("package", "main", "package name")
	sample := fs.Int("sample", 1000, "number of rows to scan for type inference")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: gen-struct [-name N] [-package P] [-sample N] file.csv")
	}
	f, err := gocsv.OpenFile(fs.Arg(0), gocsv.WithHeader(true))
	if err != nil {
		return err
	}
	headers := f.Headers()
	if len(headers) == 0 {
		return fmt.Errorf("file has no header row")
	}
	rows, err := f.GetRows()
	if err != nil {
		return err
	}
	limit := *sample
	if limit > len(rows) {
		limit = len(rows)
	}
	types := inferTypes(rows[:limit], len(headers))

	fmt.Printf("package %s\n\n", *pkg)
	if needsTimeImport(types) {
		fmt.Println("import \"time\"")
		fmt.Println()
	}
	fmt.Printf("type %s struct {\n", *name)
	maxField := 0
	fields := make([]string, len(headers))
	for i, h := range headers {
		fields[i] = exportedName(h)
		if len(fields[i]) > maxField {
			maxField = len(fields[i])
		}
	}
	maxType := 0
	for _, t := range types {
		if len(t) > maxType {
			maxType = len(t)
		}
	}
	for i, h := range headers {
		fmt.Printf("\t%-*s %-*s `csv:%q`\n", maxField, fields[i], maxType, types[i], h)
	}
	fmt.Println("}")
	_ = os.Stderr
	return nil
}

func inferTypes(rows [][]string, numCols int) []string {
	isInt := make([]bool, numCols)
	isFloat := make([]bool, numCols)
	isBool := make([]bool, numCols)
	isDate := make([]bool, numCols)
	hasData := make([]bool, numCols)
	for i := range isInt {
		isInt[i] = true
		isFloat[i] = true
		isBool[i] = true
		isDate[i] = true
	}
	for _, row := range rows {
		for col := 0; col < numCols; col++ {
			var v string
			if col < len(row) {
				v = row[col]
			}
			if v == "" {
				continue
			}
			hasData[col] = true
			if _, err := strconv.ParseInt(v, 10, 64); err != nil {
				isInt[col] = false
			}
			if _, err := strconv.ParseFloat(v, 64); err != nil {
				isFloat[col] = false
			}
			if _, err := strconv.ParseBool(v); err != nil {
				isBool[col] = false
			}
			if !looksLikeDate(v) {
				isDate[col] = false
			}
		}
	}
	out := make([]string, numCols)
	for i := 0; i < numCols; i++ {
		switch {
		case !hasData[i]:
			out[i] = "string"
		case isInt[i]:
			out[i] = "int64"
		case isFloat[i]:
			out[i] = "float64"
		case isBool[i]:
			out[i] = "bool"
		case isDate[i]:
			out[i] = "time.Time"
		default:
			out[i] = "string"
		}
	}
	return out
}

func looksLikeDate(v string) bool {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02/01/2006",
		"01/02/2006",
	}
	for _, f := range formats {
		if _, err := time.Parse(f, v); err == nil {
			return true
		}
	}
	return false
}

func needsTimeImport(types []string) bool {
	for _, t := range types {
		if t == "time.Time" {
			return true
		}
	}
	return false
}

func exportedName(header string) string {
	if header == "" {
		return "Field"
	}
	var b strings.Builder
	capitalize := true
	for _, r := range header {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			capitalize = true
			continue
		}
		if capitalize {
			b.WriteRune(unicode.ToUpper(r))
			capitalize = false
		} else {
			b.WriteRune(r)
		}
	}
	s := b.String()
	if s == "" || unicode.IsDigit(rune(s[0])) {
		s = "F" + s
	}
	return s
}
