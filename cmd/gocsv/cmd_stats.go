package main

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	gocsv "github.com/mukbeast4/go-csv"
)

func cmdStats(args []string) error {
	fs := flag.NewFlagSet("stats", flag.ContinueOnError)
	noHeader := fs.Bool("no-header", false, "no header row in input")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: stats file.csv")
	}
	f, err := gocsv.OpenFile(fs.Arg(0), gocsv.WithHeader(!*noHeader))
	if err != nil {
		return err
	}
	fmt.Printf("File: %s\n", fs.Arg(0))
	fmt.Printf("Rows: %d\n", f.RowCount())
	fmt.Printf("Cols: %d\n", f.ColCount())
	fmt.Printf("Has header: %v\n", f.HasHeader())
	if f.HasHeader() {
		fmt.Println()
		fmt.Printf("%-20s %-10s %-10s %-10s\n", "Column", "Type", "Nulls", "Distinct")
		fmt.Println(strings.Repeat("-", 55))
		headers := f.Headers()
		rows, _ := f.GetRows()
		for i, h := range headers {
			t, nulls, distinct := analyzeCol(rows, i)
			fmt.Printf("%-20s %-10s %-10d %-10d\n", h, t, nulls, distinct)
		}
	}
	return nil
}

func analyzeCol(rows [][]string, idx int) (string, int, int) {
	var nulls int
	var isInt, isFloat, isBool = true, true, true
	distinct := make(map[string]struct{})
	for _, r := range rows {
		var v string
		if idx < len(r) {
			v = r[idx]
		}
		if v == "" {
			nulls++
			continue
		}
		distinct[v] = struct{}{}
		if _, err := strconv.ParseInt(v, 10, 64); err != nil {
			isInt = false
		}
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			isFloat = false
		}
		if _, err := strconv.ParseBool(v); err != nil {
			isBool = false
		}
	}
	t := "string"
	switch {
	case isInt:
		t = "int"
	case isFloat:
		t = "float"
	case isBool:
		t = "bool"
	}
	return t, nulls, len(distinct)
}
