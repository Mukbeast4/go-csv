package gocsv

import (
	"bytes"
	"encoding/csv"
	"io"
	"os"

	"github.com/mukbeast4/go-csv/internal/encoding"
	"github.com/mukbeast4/go-csv/internal/parser"
)

type File struct {
	rows        [][]string
	headers     []string
	hasHeader   bool
	cfg         *config
	path        string
	closed      bool
	parseErrors []*ParseError
}

func NewFile(opts ...Option) *File {
	cfg := applyOptions(opts)
	return &File{cfg: cfg, hasHeader: cfg.hasHeader}
}

func OpenFile(path string, opts ...Option) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f, err := OpenBytes(data, opts...)
	if err != nil {
		return nil, err
	}
	f.path = path
	return f, nil
}

func OpenBytes(data []byte, opts ...Option) (*File, error) {
	cfg := applyOptions(opts)
	return openInternal(bytes.NewReader(data), data, cfg)
}

func OpenReader(r io.Reader, opts ...Option) (*File, error) {
	cfg := applyOptions(opts)
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return openInternal(bytes.NewReader(data), data, cfg)
}

func openInternal(reader io.Reader, raw []byte, cfg *config) (*File, error) {
	enc := cfg.encoding
	var offset int
	if enc == EncodingAuto {
		enc, offset = encoding.DetectBOM(raw)
	} else {
		_, offset = encoding.DetectBOM(raw)
	}
	payload := raw[offset:]

	if cfg.autoSniff {
		delim := sniffDelimiter(payload)
		if delim != 0 {
			cfg.dialect.Delimiter = delim
		}
	}

	var rows [][]string
	var errs []*ParseError
	var err error
	if !cfg.stdlibParser && (enc == EncodingUTF8 || enc == EncodingAuto) {
		if cfg.parallelWorkers != 1 && len(payload) >= cfg.parallelThreshold {
			rows, errs, err = parseBytesParallelAdapter(payload, cfg)
		} else {
			rows, errs, err = parseBytesAdapter(payload, cfg)
		}
	} else {
		decoded := encoding.NewDecoder(bytes.NewReader(payload), enc)
		rows, errs, err = readAll(decoded, cfg)
	}
	if err != nil && cfg.dialect.ErrorMode == ErrorModeStrict {
		return nil, err
	}

	f := &File{
		rows:        rows,
		cfg:         cfg,
		parseErrors: errs,
	}

	if cfg.headerSet {
		f.hasHeader = cfg.hasHeader
	} else if sniffHeader(rows) {
		f.hasHeader = true
	}

	if f.hasHeader && len(rows) > 0 {
		f.headers = rows[0]
		f.rows = rows[1:]
	}

	return f, nil
}

func parseBytesAdapter(data []byte, cfg *config) ([][]string, []*ParseError, error) {
	rawRows, rawErrs, err := parser.ParseBytes(data, cfg.dialect, cfg.unsafeStrings)
	if err != nil {
		pe := &ParseError{Err: err}
		if parseErr, ok := err.(*parser.ParseError); ok {
			pe = &ParseError{Line: parseErr.Line, Column: parseErr.Column, Offset: parseErr.Offset, Err: parseErr.Err}
		}
		return rawRows, convertParseErrors(rawErrs), pe
	}
	return rawRows, convertParseErrors(rawErrs), nil
}

func convertParseErrors(src []*parser.ParseError) []*ParseError {
	if src == nil {
		return nil
	}
	out := make([]*ParseError, len(src))
	for i, e := range src {
		out[i] = &ParseError{Line: e.Line, Column: e.Column, Offset: e.Offset, Err: e.Err}
	}
	return out
}

func parseBytesParallelAdapter(data []byte, cfg *config) ([][]string, []*ParseError, error) {
	rawRows, rawErrs, err := parser.ParseBytesParallel(data, cfg.dialect, cfg.unsafeStrings, cfg.parallelWorkers)
	if err != nil {
		pe := &ParseError{Err: err}
		if parseErr, ok := err.(*parser.ParseError); ok {
			pe = &ParseError{Line: parseErr.Line, Column: parseErr.Column, Offset: parseErr.Offset, Err: parseErr.Err}
		}
		return rawRows, convertParseErrors(rawErrs), pe
	}
	return rawRows, convertParseErrors(rawErrs), nil
}

func readAll(r io.Reader, cfg *config) ([][]string, []*ParseError, error) {
	if cfg.stdlibParser {
		cr := csv.NewReader(r)
		cr.Comma = cfg.dialect.Delimiter
		if cfg.dialect.Comment != 0 {
			cr.Comment = cfg.dialect.Comment
		}
		cr.LazyQuotes = cfg.dialect.LazyQuotes
		cr.TrimLeadingSpace = cfg.dialect.TrimLeadingSpace
		cr.FieldsPerRecord = cfg.dialect.FieldsPerRecord
		rows, err := cr.ReadAll()
		return rows, nil, err
	}

	p := parser.New(r, cfg.dialect)
	var rows [][]string
	var errs []*ParseError

	for {
		row, err := p.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			pe := convertParseError(err, p)
			switch cfg.dialect.ErrorMode {
			case ErrorModeSkip:
				errs = append(errs, pe)
				continue
			case ErrorModeCollect:
				errs = append(errs, pe)
				if row != nil {
					rows = append(rows, row)
				}
				continue
			default:
				return rows, errs, pe
			}
		}
		rows = append(rows, row)
	}
	return rows, errs, nil
}

func convertParseError(err error, p *parser.Parser) *ParseError {
	if pe, ok := err.(*parser.ParseError); ok {
		return &ParseError{
			Line:   pe.Line,
			Column: pe.Column,
			Offset: pe.Offset,
			Err:    pe.Err,
		}
	}
	return &ParseError{
		Line:   p.Line(),
		Offset: p.Offset(),
		Err:    err,
	}
}

func (f *File) RowCount() int {
	if f.closed {
		return 0
	}
	return len(f.rows)
}

func (f *File) ColCount() int {
	if f.closed {
		return 0
	}
	if len(f.headers) > 0 {
		return len(f.headers)
	}
	n := 0
	for _, row := range f.rows {
		if len(row) > n {
			n = len(row)
		}
	}
	return n
}

func (f *File) Dimension() string {
	rows := f.RowCount()
	cols := f.ColCount()
	if rows == 0 || cols == 0 {
		return ""
	}
	if f.hasHeader {
		rows++
	}
	return Cells(1, 1, cols, rows)
}

func (f *File) HasHeader() bool {
	return f.hasHeader
}

func (f *File) ParseErrors() []*ParseError {
	return f.parseErrors
}

func (f *File) Path() string {
	return f.path
}

func (f *File) checkClosed() error {
	if f.closed {
		return ErrFileClosed
	}
	return nil
}
