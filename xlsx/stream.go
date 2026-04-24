package xlsx

import (
	"fmt"
	"io"
	"os"

	gocsv "github.com/mukbeast4/go-csv"
	"github.com/xuri/excelize/v2"
)

// NewStreamReader returns a row iterator reading from an XLSX file without
// loading all rows in memory. It uses excelize's row iterator internally.
func NewStreamReader(path, sheetName string, opts ...Option) (*gocsv.RowIterator, io.Closer, error) {
	cfg := apply(opts)
	if sheetName != "" {
		cfg.sheetName = sheetName
	}
	xl, err := excelize.OpenFile(path)
	if err != nil {
		return nil, nil, err
	}
	rows, err := xl.Rows(cfg.sheetName)
	if err != nil {
		xl.Close()
		return nil, nil, err
	}

	var headers []string
	if cfg.hasHeader && rows.Next() {
		headers, err = rows.Columns()
		if err != nil {
			rows.Close()
			xl.Close()
			return nil, nil, err
		}
	}

	next := func() ([]string, error) {
		if !rows.Next() {
			if err := rows.Error(); err != nil {
				return nil, err
			}
			return nil, io.EOF
		}
		return rows.Columns()
	}

	closer := &xlsxReadCloser{xl: xl, rows: rows}
	return gocsv.NewRowIteratorFromFunc(headers, next, closer), closer, nil
}

type xlsxReadCloser struct {
	xl   *excelize.File
	rows *excelize.Rows
}

func (c *xlsxReadCloser) Close() error {
	if c.rows != nil {
		c.rows.Close()
	}
	if c.xl != nil {
		return c.xl.Close()
	}
	return nil
}

// StreamWriter writes rows to an XLSX sheet incrementally. Unlike
// *gocsv.StreamWriter it is backed by excelize's native stream writer, which
// avoids buffering all rows before flushing.
type StreamWriter struct {
	xl     *excelize.File
	sw     *excelize.StreamWriter
	path   string
	rowIdx int
	cfg    *config
}

func NewStreamWriter(path, sheetName string, opts ...Option) (*StreamWriter, io.Closer, error) {
	cfg := apply(opts)
	if sheetName != "" {
		cfg.sheetName = sheetName
	}
	xl := excelize.NewFile()
	if cfg.sheetName != "Sheet1" {
		xl.SetSheetName("Sheet1", cfg.sheetName)
	}
	sw, err := xl.NewStreamWriter(cfg.sheetName)
	if err != nil {
		xl.Close()
		return nil, nil, err
	}
	w := &StreamWriter{xl: xl, sw: sw, path: path, rowIdx: 1, cfg: cfg}
	return w, w, nil
}

func (w *StreamWriter) WriteHeader(headers []string) error {
	cell, err := excelize.CoordinatesToCellName(1, w.rowIdx)
	if err != nil {
		return err
	}
	values := make([]any, len(headers))
	for i, h := range headers {
		values[i] = h
	}
	if err := w.sw.SetRow(cell, values); err != nil {
		return err
	}
	w.rowIdx++
	return nil
}

func (w *StreamWriter) WriteStrRow(row []string) error {
	cell, err := excelize.CoordinatesToCellName(1, w.rowIdx)
	if err != nil {
		return err
	}
	values := toTypedRow(row, w.cfg)
	if err := w.sw.SetRow(cell, values); err != nil {
		return err
	}
	w.rowIdx++
	return nil
}

func (w *StreamWriter) WriteRow(values []any) error {
	cell, err := excelize.CoordinatesToCellName(1, w.rowIdx)
	if err != nil {
		return err
	}
	if err := w.sw.SetRow(cell, values); err != nil {
		return err
	}
	w.rowIdx++
	return nil
}

func (w *StreamWriter) WriteRecord(record map[string]any, headers []string) error {
	values := make([]any, len(headers))
	for i, h := range headers {
		values[i] = record[h]
	}
	return w.WriteRow(values)
}

func (w *StreamWriter) Flush() error {
	return w.sw.Flush()
}

func (w *StreamWriter) Close() error {
	if err := w.sw.Flush(); err != nil {
		w.xl.Close()
		return fmt.Errorf("xlsx: flush: %w", err)
	}
	if w.path == "" {
		return w.xl.Close()
	}
	out, err := os.Create(w.path)
	if err != nil {
		w.xl.Close()
		return err
	}
	if err := w.xl.Write(out); err != nil {
		out.Close()
		w.xl.Close()
		return err
	}
	if err := out.Close(); err != nil {
		w.xl.Close()
		return err
	}
	return w.xl.Close()
}
