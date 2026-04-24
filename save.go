package gocsv

import (
	"bytes"
	"io"
	"os"

	"github.com/mukbeast4/go-csv/internal/encoding"
	"github.com/mukbeast4/go-csv/internal/parser"
)

func (f *File) Save() error {
	if err := f.checkClosed(); err != nil {
		return err
	}
	if f.path == "" {
		return os.ErrNotExist
	}
	return f.SaveAs(f.path)
}

func (f *File) SaveAs(path string) error {
	if err := f.checkClosed(); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := f.Write(file); err != nil {
		return err
	}
	f.path = path
	return nil
}

func (f *File) Write(w io.Writer) error {
	if err := f.checkClosed(); err != nil {
		return err
	}
	enc := f.cfg.encoding
	if enc == EncodingAuto {
		enc = EncodingUTF8
	}
	sink := encoding.NewEncoder(w, enc, f.cfg.writeBOM)
	pw := parser.NewWriter(sink, f.cfg.dialect)

	if f.hasHeader && len(f.headers) > 0 {
		if err := pw.WriteRow(f.headers); err != nil {
			return err
		}
	}
	for _, row := range f.rows {
		if err := pw.WriteRow(row); err != nil {
			return err
		}
	}
	return pw.Flush()
}

func (f *File) WriteToBuffer() (*bytes.Buffer, error) {
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return &buf, nil
}

func (f *File) WriteToFile(path string) error {
	return f.SaveAs(path)
}

func (f *File) Close() error {
	if f.closed {
		return nil
	}
	f.closed = true
	f.rows = nil
	f.headers = nil
	return nil
}
