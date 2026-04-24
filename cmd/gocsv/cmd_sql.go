package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	gocsv "github.com/mukbeast4/go-csv"
	"github.com/mukbeast4/go-csv/cmd/gocsv/internal/sql"
)

func cmdSQL(args []string) error {
	fs := flag.NewFlagSet("sql", flag.ContinueOnError)
	query := fs.String("q", "", "SQL query (alternative to positional arg)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	var queryStr string
	var filePath string
	switch {
	case *query != "" && fs.NArg() >= 1:
		queryStr = *query
		filePath = fs.Arg(0)
	case *query != "":
		return fmt.Errorf("usage: sql -q \"SELECT ...\" file.csv")
	case fs.NArg() >= 2:
		queryStr = fs.Arg(0)
		filePath = fs.Arg(1)
	default:
		return fmt.Errorf(`usage: sql "SELECT col1, col2 FROM t WHERE col op value GROUP BY col ORDER BY col LIMIT n" file.csv`)
	}
	queryStr = strings.TrimSpace(queryStr)
	stmt, err := sql.Parse(queryStr)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	f, err := gocsv.OpenFile(filePath, gocsv.WithHeader(true))
	if err != nil {
		return err
	}
	out, err := sql.Execute(stmt, f)
	if err != nil {
		return err
	}
	return out.Write(os.Stdout)
}
