package ods

import (
	"os"

	gocsv "github.com/mukbeast4/go-csv"
	goods "github.com/mukbeast4/go-ods"
)

type Option func(*config)

type config struct {
	sheetName string
	hasHeader bool
}

func defaultConfig() *config {
	return &config{sheetName: "Sheet1", hasHeader: true}
}

func WithSheetName(name string) Option {
	return func(c *config) { c.sheetName = name }
}

func WithHeader(enabled bool) Option {
	return func(c *config) { c.hasHeader = enabled }
}

func ToODS(f *gocsv.File, path string, opts ...Option) error {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	ods := goods.NewFile()
	ods.SetSheetName("Sheet1", cfg.sheetName)
	return writeSheet(ods, cfg.sheetName, f, cfg.hasHeader, path)
}

func AppendSheet(odsPath, sheetName string, f *gocsv.File, opts ...Option) error {
	cfg := defaultConfig()
	cfg.sheetName = sheetName
	for _, opt := range opts {
		opt(cfg)
	}
	var ods *goods.File
	if _, err := os.Stat(odsPath); err == nil {
		ods, err = goods.OpenFile(odsPath)
		if err != nil {
			return err
		}
	} else {
		ods = goods.NewFile()
		ods.SetSheetName("Sheet1", sheetName)
	}
	if !containsSheet(ods.GetSheetList(), sheetName) {
		if _, err := ods.NewSheet(sheetName); err != nil {
			return err
		}
	}
	return writeSheet(ods, sheetName, f, cfg.hasHeader, odsPath)
}

func FromODS(path, sheetName string, opts ...Option) (*gocsv.File, error) {
	ods, err := goods.OpenFile(path)
	if err != nil {
		return nil, err
	}
	cfg := defaultConfig()
	cfg.sheetName = sheetName
	for _, opt := range opts {
		opt(cfg)
	}
	rows, err := ods.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	out := gocsv.NewFile()
	if cfg.hasHeader && len(rows) > 0 {
		out.SetHeaders(rows[0])
		for _, row := range rows[1:] {
			out.AppendStrRow(row)
		}
	} else {
		for _, row := range rows {
			out.AppendStrRow(row)
		}
	}
	return out, nil
}

func writeSheet(ods *goods.File, sheetName string, f *gocsv.File, withHeader bool, path string) error {
	rowIdx := 1
	if withHeader {
		headers := f.Headers()
		if len(headers) > 0 {
			values := toAnySlice(headers)
			if err := ods.SetRowValues(sheetName, rowIdx, values); err != nil {
				return err
			}
			rowIdx++
		}
	}
	rows, err := f.GetRows()
	if err != nil {
		return err
	}
	for _, r := range rows {
		values := toAnySlice(r)
		if err := ods.SetRowValues(sheetName, rowIdx, values); err != nil {
			return err
		}
		rowIdx++
	}
	return ods.SaveAs(path)
}

func toAnySlice(row []string) []any {
	out := make([]any, len(row))
	for i, v := range row {
		out[i] = v
	}
	return out
}

func containsSheet(list []string, name string) bool {
	for _, s := range list {
		if s == name {
			return true
		}
	}
	return false
}
