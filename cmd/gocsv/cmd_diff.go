package main

import (
	"flag"
	"fmt"
	"strings"

	gocsv "github.com/mukbeast4/go-csv"
	"github.com/mukbeast4/go-csv/merge"
)

func cmdDiff(args []string) error {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	on := fs.String("on", "", "column to match rows on")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *on == "" || fs.NArg() < 2 {
		return fmt.Errorf("usage: diff -on col before.csv after.csv")
	}
	before, err := gocsv.OpenFile(fs.Arg(0), gocsv.WithHeader(true))
	if err != nil {
		return err
	}
	after, err := gocsv.OpenFile(fs.Arg(1), gocsv.WithHeader(true))
	if err != nil {
		return err
	}
	d, err := merge.Diff(before, after, merge.On(*on))
	if err != nil {
		return err
	}
	fmt.Printf("# Added: %d\n", len(d.Added))
	for _, r := range d.Added {
		fmt.Printf("+ %s\n", strings.Join(r, ","))
	}
	fmt.Printf("\n# Removed: %d\n", len(d.Removed))
	for _, r := range d.Removed {
		fmt.Printf("- %s\n", strings.Join(r, ","))
	}
	fmt.Printf("\n# Modified: %d\n", len(d.Modified))
	for _, m := range d.Modified {
		fmt.Printf("~ key=%s fields=%v\n", m.Key, m.Fields)
		fmt.Printf("    before: %s\n", strings.Join(m.Before, ","))
		fmt.Printf("    after:  %s\n", strings.Join(m.After, ","))
	}
	return nil
}
