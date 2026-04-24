package main

import (
	"flag"
	"fmt"
	"os"

	gocsv "github.com/mukbeast4/go-csv"
	"github.com/mukbeast4/go-csv/merge"
)

func cmdJoin(args []string) error {
	fs := flag.NewFlagSet("join", flag.ContinueOnError)
	on := fs.String("on", "", "column to join on")
	mode := fs.String("mode", "inner", "join mode: inner, left, right, full")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *on == "" || fs.NArg() < 2 {
		return fmt.Errorf("usage: join -on col [-mode inner|left|right|full] a.csv b.csv")
	}
	a, err := gocsv.OpenFile(fs.Arg(0), gocsv.WithHeader(true))
	if err != nil {
		return err
	}
	b, err := gocsv.OpenFile(fs.Arg(1), gocsv.WithHeader(true))
	if err != nil {
		return err
	}
	key := merge.On(*on)
	var out *gocsv.File
	switch *mode {
	case "inner":
		out, err = merge.InnerJoin(a, b, key)
	case "left":
		out, err = merge.LeftJoin(a, b, key)
	case "right":
		out, err = merge.RightJoin(a, b, key)
	case "full":
		out, err = merge.FullJoin(a, b, key)
	default:
		return fmt.Errorf("unknown mode: %s", *mode)
	}
	if err != nil {
		return err
	}
	return out.Write(os.Stdout)
}
