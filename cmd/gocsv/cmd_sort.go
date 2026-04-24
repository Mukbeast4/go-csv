package main

import (
	"flag"
	"fmt"
	"os"

	gocsv "github.com/mukbeast4/go-csv"
	"github.com/mukbeast4/go-csv/query"
)

func cmdSort(args []string) error {
	fs := flag.NewFlagSet("sort", flag.ContinueOnError)
	col := fs.String("c", "", "column to sort by")
	desc := fs.Bool("desc", false, "sort descending")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *col == "" || fs.NArg() < 1 {
		return fmt.Errorf("usage: sort -c col [-desc] file.csv")
	}
	f, err := gocsv.OpenFile(fs.Arg(0), gocsv.WithHeader(true))
	if err != nil {
		return err
	}
	dir := query.Asc
	if *desc {
		dir = query.Desc
	}
	out := query.From(f).OrderBy(*col, dir).ToFile()
	if out == nil {
		return query.From(f).OrderBy(*col, dir).Err()
	}
	return out.Write(os.Stdout)
}
