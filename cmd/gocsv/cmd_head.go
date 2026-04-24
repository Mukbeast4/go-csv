package main

import (
	"flag"
	"fmt"
	"os"

	gocsv "github.com/mukbeast4/go-csv"
)

func cmdHead(args []string) error {
	fs := flag.NewFlagSet("head", flag.ContinueOnError)
	n := fs.Int("n", 10, "number of rows to print")
	noHeader := fs.Bool("no-header", false, "no header row in input")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: head [-n N] file.csv")
	}
	f, err := gocsv.OpenFile(fs.Arg(0), gocsv.WithHeader(!*noHeader))
	if err != nil {
		return err
	}
	sw := gocsv.NewStreamWriter(os.Stdout)
	defer sw.Close()
	if f.HasHeader() {
		sw.WriteHeader(f.Headers())
	}
	rows, _ := f.GetRows()
	limit := *n
	if limit > len(rows) {
		limit = len(rows)
	}
	for i := 0; i < limit; i++ {
		sw.WriteStrRow(rows[i])
	}
	return nil
}

func cmdTail(args []string) error {
	fs := flag.NewFlagSet("tail", flag.ContinueOnError)
	n := fs.Int("n", 10, "number of rows to print")
	noHeader := fs.Bool("no-header", false, "no header row in input")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: tail [-n N] file.csv")
	}
	f, err := gocsv.OpenFile(fs.Arg(0), gocsv.WithHeader(!*noHeader))
	if err != nil {
		return err
	}
	sw := gocsv.NewStreamWriter(os.Stdout)
	defer sw.Close()
	if f.HasHeader() {
		sw.WriteHeader(f.Headers())
	}
	rows, _ := f.GetRows()
	start := len(rows) - *n
	if start < 0 {
		start = 0
	}
	for i := start; i < len(rows); i++ {
		sw.WriteStrRow(rows[i])
	}
	return nil
}
