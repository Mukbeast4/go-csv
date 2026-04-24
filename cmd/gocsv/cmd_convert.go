package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	gocsv "github.com/mukbeast4/go-csv"
	"github.com/mukbeast4/go-csv/compress"
	"github.com/mukbeast4/go-csv/jsonx"
	"github.com/mukbeast4/go-csv/ods"
	"github.com/mukbeast4/go-csv/xlsx"
)

func cmdConvert(args []string) error {
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	to := fs.String("to", "", "target format: csv, ods, xlsx, json, ndjson, gz, bz2, zst")
	sheet := fs.String("sheet", "Sheet1", "sheet name (for ods/xlsx)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 2 || *to == "" {
		return fmt.Errorf("usage: convert -to <format> input.csv output.<ext>")
	}
	input := fs.Arg(0)
	output := fs.Arg(1)

	f, err := loadInput(input, *sheet)
	if err != nil {
		return err
	}

	switch strings.ToLower(*to) {
	case "csv":
		return f.SaveAs(output)
	case "ods":
		return ods.ToODS(f, output, ods.WithSheetName(*sheet))
	case "xlsx":
		return xlsx.ToXLSX(f, output, xlsx.WithSheetName(*sheet))
	case "json":
		out, err := os.Create(output)
		if err != nil {
			return err
		}
		defer out.Close()
		return jsonx.ToJSON(f, out, jsonx.WithPretty(true))
	case "ndjson":
		out, err := os.Create(output)
		if err != nil {
			return err
		}
		defer out.Close()
		return jsonx.ToNDJSON(f, out)
	case "gz", "gzip", "bz2", "bzip2", "zst", "zstd":
		return compress.SaveAs(f, output)
	default:
		return fmt.Errorf("unknown format: %s", *to)
	}
}

func loadInput(input, sheet string) (*gocsv.File, error) {
	lower := strings.ToLower(input)
	switch {
	case isCompressed(input):
		return compress.Open(input, gocsv.WithHeader(true))
	case strings.HasSuffix(lower, ".ods"):
		return ods.FromODS(input, sheet)
	case strings.HasSuffix(lower, ".xlsx"):
		return xlsx.FromXLSX(input, sheet)
	case strings.HasSuffix(lower, ".json"):
		file, err := os.Open(input)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return jsonx.FromJSON(file)
	case strings.HasSuffix(lower, ".ndjson"), strings.HasSuffix(lower, ".jsonl"):
		file, err := os.Open(input)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		return jsonx.FromNDJSON(file)
	default:
		return gocsv.OpenFile(input, gocsv.WithHeader(true))
	}
}

func isCompressed(path string) bool {
	lower := strings.ToLower(path)
	return strings.HasSuffix(lower, ".gz") ||
		strings.HasSuffix(lower, ".gzip") ||
		strings.HasSuffix(lower, ".bz2") ||
		strings.HasSuffix(lower, ".bzip2") ||
		strings.HasSuffix(lower, ".zst") ||
		strings.HasSuffix(lower, ".zstd")
}
