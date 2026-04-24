package xlsx

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	gocsv "github.com/mukbeast4/go-csv"
	"github.com/xuri/excelize/v2"
)

type Option func(*config)

type config struct {
	sheetName  string
	hasHeader  bool
	dateFormat string
	autoFilter bool
	typeInfer  bool
}

func defaultConfig() *config {
	return &config{
		sheetName:  "Sheet1",
		hasHeader:  true,
		dateFormat: "yyyy-mm-dd",
		typeInfer:  true,
	}
}

func WithSheetName(name string) Option    { return func(c *config) { c.sheetName = name } }
func WithHeader(enabled bool) Option      { return func(c *config) { c.hasHeader = enabled } }
func WithDateFormat(format string) Option { return func(c *config) { c.dateFormat = format } }
func WithAutoFilter(enabled bool) Option  { return func(c *config) { c.autoFilter = enabled } }
func WithTypeInfer(enabled bool) Option   { return func(c *config) { c.typeInfer = enabled } }

func ToXLSX(f *gocsv.File, path string, opts ...Option) error {
	cfg := apply(opts)
	xl := excelize.NewFile()
	defer xl.Close()
	if cfg.sheetName != "Sheet1" {
		xl.SetSheetName("Sheet1", cfg.sheetName)
	}
	if err := writeSheet(xl, cfg.sheetName, f, cfg); err != nil {
		return err
	}
	return xl.SaveAs(path)
}

func WriteXLSX(f *gocsv.File, w io.Writer, opts ...Option) error {
	cfg := apply(opts)
	xl := excelize.NewFile()
	defer xl.Close()
	if cfg.sheetName != "Sheet1" {
		xl.SetSheetName("Sheet1", cfg.sheetName)
	}
	if err := writeSheet(xl, cfg.sheetName, f, cfg); err != nil {
		return err
	}
	return xl.Write(w)
}

func FromXLSX(path, sheetName string, opts ...Option) (*gocsv.File, error) {
	cfg := apply(opts)
	if sheetName != "" {
		cfg.sheetName = sheetName
	}
	xl, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer xl.Close()
	return readSheet(xl, cfg)
}

func ReadXLSX(r io.Reader, sheetName string, opts ...Option) (*gocsv.File, error) {
	cfg := apply(opts)
	if sheetName != "" {
		cfg.sheetName = sheetName
	}
	xl, err := excelize.OpenReader(r)
	if err != nil {
		return nil, err
	}
	defer xl.Close()
	return readSheet(xl, cfg)
}

func AppendSheet(xlsxPath, sheetName string, f *gocsv.File, opts ...Option) error {
	cfg := apply(opts)
	cfg.sheetName = sheetName

	var xl *excelize.File
	var err error
	if _, statErr := os.Stat(xlsxPath); statErr == nil {
		xl, err = excelize.OpenFile(xlsxPath)
		if err != nil {
			return err
		}
	} else {
		xl = excelize.NewFile()
	}
	defer xl.Close()

	if _, err := xl.GetSheetIndex(sheetName); err != nil || !sheetExists(xl, sheetName) {
		if _, err := xl.NewSheet(sheetName); err != nil {
			return err
		}
	}
	if err := writeSheet(xl, sheetName, f, cfg); err != nil {
		return err
	}
	return xl.SaveAs(xlsxPath)
}

func SheetNames(path string) ([]string, error) {
	xl, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer xl.Close()
	return xl.GetSheetList(), nil
}

func apply(opts []Option) *config {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func sheetExists(xl *excelize.File, name string) bool {
	for _, s := range xl.GetSheetList() {
		if s == name {
			return true
		}
	}
	return false
}

func writeSheet(xl *excelize.File, sheet string, f *gocsv.File, cfg *config) error {
	rowIdx := 1
	if cfg.hasHeader {
		headers := f.Headers()
		if len(headers) > 0 {
			cell, err := excelize.CoordinatesToCellName(1, rowIdx)
			if err != nil {
				return err
			}
			if err := xl.SetSheetRow(sheet, cell, &headers); err != nil {
				return err
			}
			if cfg.autoFilter {
				lastCol, err := excelize.CoordinatesToCellName(len(headers), rowIdx)
				if err != nil {
					return err
				}
				if err := xl.AutoFilter(sheet, fmt.Sprintf("%s:%s", cell, lastCol), []excelize.AutoFilterOptions{}); err != nil {
					return err
				}
			}
			rowIdx++
		}
	}
	rows, err := f.GetRows()
	if err != nil {
		return err
	}
	for _, row := range rows {
		cell, err := excelize.CoordinatesToCellName(1, rowIdx)
		if err != nil {
			return err
		}
		values := toTypedRow(row, cfg)
		if err := xl.SetSheetRow(sheet, cell, &values); err != nil {
			return err
		}
		rowIdx++
	}
	return nil
}

func toTypedRow(row []string, cfg *config) []any {
	out := make([]any, len(row))
	for i, v := range row {
		if !cfg.typeInfer {
			out[i] = v
			continue
		}
		out[i] = typed(v)
	}
	return out
}

func typed(v string) any {
	if v == "" {
		return v
	}
	if n, err := strconv.ParseInt(v, 10, 64); err == nil {
		return n
	}
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return f
	}
	lower := strings.ToLower(v)
	if lower == "true" {
		return true
	}
	if lower == "false" {
		return false
	}
	return v
}

func readSheet(xl *excelize.File, cfg *config) (*gocsv.File, error) {
	rows, err := xl.GetRows(cfg.sheetName)
	if err != nil {
		return nil, err
	}
	out := gocsv.NewFile()
	if cfg.hasHeader && len(rows) > 0 {
		out.SetHeaders(rows[0])
		for _, row := range rows[1:] {
			if err := out.AppendStrRow(row); err != nil {
				return nil, err
			}
		}
	} else {
		for _, row := range rows {
			if err := out.AppendStrRow(row); err != nil {
				return nil, err
			}
		}
	}
	return out, nil
}
