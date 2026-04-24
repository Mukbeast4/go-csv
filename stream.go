package gocsv

import (
	"io"
	"os"

	"github.com/mukbeast4/go-csv/internal/encoding"
	"github.com/mukbeast4/go-csv/internal/parser"
)

type StreamWriter struct {
	writer    *parser.Writer
	cfg       *config
	headers   []string
	hasHeader bool
	closed    bool
	closer    io.Closer
}

func NewStreamWriter(w io.Writer, opts ...Option) *StreamWriter {
	cfg := applyOptions(opts)
	enc := cfg.encoding
	if enc == EncodingAuto {
		enc = EncodingUTF8
	}
	sink := encoding.NewEncoder(w, enc, cfg.writeBOM)
	return &StreamWriter{
		writer: parser.NewWriter(sink, cfg.dialect),
		cfg:    cfg,
	}
}

func NewStreamWriterToFile(path string, opts ...Option) (*StreamWriter, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	sw := NewStreamWriter(file, opts...)
	sw.closer = file
	return sw, nil
}

func (sw *StreamWriter) WriteHeader(headers []string) error {
	if sw.closed {
		return ErrStreamClosed
	}
	if sw.hasHeader {
		return nil
	}
	cp := make([]string, len(headers))
	copy(cp, headers)
	sw.headers = cp
	sw.hasHeader = true
	return sw.writer.WriteRow(cp)
}

func (sw *StreamWriter) WriteRow(values []any) error {
	if sw.closed {
		return ErrStreamClosed
	}
	strs := make([]string, len(values))
	for i, v := range values {
		_, s := detectCellType(v)
		strs[i] = s
	}
	return sw.writer.WriteRow(strs)
}

func (sw *StreamWriter) WriteStrRow(values []string) error {
	if sw.closed {
		return ErrStreamClosed
	}
	return sw.writer.WriteRow(values)
}

func (sw *StreamWriter) WriteRecord(record map[string]any) error {
	if sw.closed {
		return ErrStreamClosed
	}
	if !sw.hasHeader {
		return ErrNoHeader
	}
	row := make([]string, len(sw.headers))
	for i, h := range sw.headers {
		if v, ok := record[h]; ok {
			_, s := detectCellType(v)
			row[i] = s
		}
	}
	return sw.writer.WriteRow(row)
}

func (sw *StreamWriter) Flush() error {
	if sw.closed {
		return ErrStreamClosed
	}
	return sw.writer.Flush()
}

func (sw *StreamWriter) Close() error {
	if sw.closed {
		return nil
	}
	sw.closed = true
	if err := sw.writer.Flush(); err != nil {
		if sw.closer != nil {
			sw.closer.Close()
		}
		return err
	}
	if sw.closer != nil {
		return sw.closer.Close()
	}
	return nil
}
