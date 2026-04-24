package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	gocsv "github.com/mukbeast4/go-csv"
	"github.com/mukbeast4/go-csv/query"
)

func cmdSelect(args []string) error {
	fs := flag.NewFlagSet("select", flag.ContinueOnError)
	cols := fs.String("c", "", "comma-separated list of columns")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *cols == "" || fs.NArg() < 1 {
		return fmt.Errorf("usage: select -c col1,col2 file.csv")
	}
	f, err := gocsv.OpenFile(fs.Arg(0), gocsv.WithHeader(true))
	if err != nil {
		return err
	}
	columns := strings.Split(*cols, ",")
	for i, c := range columns {
		columns[i] = strings.TrimSpace(c)
	}
	out := query.From(f).Select(columns...).ToFile()
	if out == nil {
		return query.From(f).Select(columns...).Err()
	}
	return out.Write(os.Stdout)
}
