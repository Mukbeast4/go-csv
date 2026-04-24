package main

import (
	"flag"
	"fmt"
	"strings"

	gocsv "github.com/mukbeast4/go-csv"
	"github.com/mukbeast4/go-csv/compress"
	"github.com/mukbeast4/go-csv/ods"
)

func cmdConvert(args []string) error {
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	to := fs.String("to", "", "target format: csv, ods, gz, bz2")
	sheet := fs.String("sheet", "Sheet1", "sheet name (for ods)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 2 || *to == "" {
		return fmt.Errorf("usage: convert -to <format> input.csv output.<ext>")
	}
	input := fs.Arg(0)
	output := fs.Arg(1)

	var f *gocsv.File
	var err error
	if isCompressed(input) {
		f, err = compress.Open(input, gocsv.WithHeader(true))
	} else if strings.HasSuffix(strings.ToLower(input), ".ods") {
		f, err = ods.FromODS(input, *sheet)
	} else {
		f, err = gocsv.OpenFile(input, gocsv.WithHeader(true))
	}
	if err != nil {
		return err
	}

	switch strings.ToLower(*to) {
	case "csv":
		return f.SaveAs(output)
	case "ods":
		return ods.ToODS(f, output, ods.WithSheetName(*sheet))
	case "gz", "gzip":
		return compress.SaveAs(f, output)
	default:
		return fmt.Errorf("unknown format: %s", *to)
	}
}

func isCompressed(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".gz") || strings.HasSuffix(lower, ".bz2")
}
