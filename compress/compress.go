package compress

import (
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"

	gocsv "github.com/mukbeast4/go-csv"
)

func Open(path string, opts ...gocsv.Option) (*gocsv.File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	format := DetectFormat(path)
	decoded, err := decode(data, format)
	if err != nil {
		return nil, err
	}
	return gocsv.OpenBytes(decoded, opts...)
}

func SaveAs(f *gocsv.File, path string, opts ...gocsv.Option) error {
	format := DetectFormat(path)
	var payload bytes.Buffer
	if err := f.Write(&payload); err != nil {
		return err
	}
	encoded, err := encode(payload.Bytes(), format)
	if err != nil {
		return err
	}
	return os.WriteFile(path, encoded, 0644)
}

func NewStreamReader(path string, opts ...gocsv.Option) (*gocsv.RowIterator, io.Closer, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	format := DetectFormat(path)
	reader, closer, err := wrapDecoder(file, format)
	if err != nil {
		file.Close()
		return nil, nil, err
	}
	it, err := gocsv.StreamReader(reader, opts...)
	if err != nil {
		closer.Close()
		return nil, nil, err
	}
	return it, combinedCloser{it, closer, file}, nil
}

type combinedCloser struct {
	it     *gocsv.RowIterator
	wrap   io.Closer
	origin io.Closer
}

func (c combinedCloser) Close() error {
	c.it.Close()
	c.wrap.Close()
	return c.origin.Close()
}

func NewStreamWriter(path string, opts ...gocsv.Option) (*gocsv.StreamWriter, io.Closer, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	format := DetectFormat(path)
	writer, closer, err := wrapEncoder(file, format)
	if err != nil {
		file.Close()
		return nil, nil, err
	}
	sw := gocsv.NewStreamWriter(writer, opts...)
	return sw, compositeCloser{sw: sw, wrap: closer, origin: file}, nil
}

type compositeCloser struct {
	sw     *gocsv.StreamWriter
	wrap   io.Closer
	origin io.Closer
}

func (c compositeCloser) Close() error {
	if err := c.sw.Close(); err != nil {
		c.wrap.Close()
		c.origin.Close()
		return err
	}
	if err := c.wrap.Close(); err != nil {
		c.origin.Close()
		return err
	}
	return c.origin.Close()
}

func decode(data []byte, format Format) ([]byte, error) {
	switch format {
	case FormatNone:
		return data, nil
	case FormatGzip:
		gr, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		defer gr.Close()
		return io.ReadAll(gr)
	case FormatBzip2:
		br := bzip2.NewReader(bytes.NewReader(data))
		return io.ReadAll(br)
	default:
		return nil, fmt.Errorf("compress: unsupported format")
	}
}

func encode(data []byte, format Format) ([]byte, error) {
	switch format {
	case FormatNone:
		return data, nil
	case FormatGzip:
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		if _, err := gw.Write(data); err != nil {
			return nil, err
		}
		if err := gw.Close(); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	case FormatBzip2:
		return nil, fmt.Errorf("compress: bzip2 write not supported in stdlib")
	default:
		return nil, fmt.Errorf("compress: unsupported format")
	}
}

type nopCloser struct{}

func (nopCloser) Close() error { return nil }

func wrapDecoder(r io.Reader, format Format) (io.Reader, io.Closer, error) {
	switch format {
	case FormatNone:
		return r, nopCloser{}, nil
	case FormatGzip:
		gr, err := gzip.NewReader(r)
		if err != nil {
			return nil, nil, err
		}
		return gr, gr, nil
	case FormatBzip2:
		return bzip2.NewReader(r), nopCloser{}, nil
	default:
		return nil, nil, fmt.Errorf("compress: unsupported format")
	}
}

func wrapEncoder(w io.Writer, format Format) (io.Writer, io.Closer, error) {
	switch format {
	case FormatNone:
		return w, nopCloser{}, nil
	case FormatGzip:
		gw := gzip.NewWriter(w)
		return gw, gw, nil
	case FormatBzip2:
		return nil, nil, fmt.Errorf("compress: bzip2 write not supported")
	default:
		return nil, nil, fmt.Errorf("compress: unsupported format")
	}
}
