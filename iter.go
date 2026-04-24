package gocsv

import (
	"bytes"
	"io"
	"os"

	"github.com/mukbeast4/go-csv/internal/encoding"
	"github.com/mukbeast4/go-csv/internal/parser"
)

type RowIterator struct {
	parser    *parser.Parser
	cfg       *config
	rows      [][]string
	idx       int
	current   []string
	rowIdx    int
	err       error
	closed    bool
	streaming bool
	headers   []string
	closer    io.Closer
}

func (f *File) NewRowIterator() (*RowIterator, error) {
	if err := f.checkClosed(); err != nil {
		return nil, err
	}
	cp := make([][]string, len(f.rows))
	for i, r := range f.rows {
		rowCopy := make([]string, len(r))
		copy(rowCopy, r)
		cp[i] = rowCopy
	}
	headers := make([]string, len(f.headers))
	copy(headers, f.headers)
	return &RowIterator{
		rows:    cp,
		headers: headers,
		cfg:     f.cfg,
		idx:     -1,
		rowIdx:  -1,
	}, nil
}

func StreamReader(r io.Reader, opts ...Option) (*RowIterator, error) {
	cfg := applyOptions(opts)
	data, err := peekSample(r, 8192)
	if err != nil {
		return nil, err
	}
	enc := cfg.encoding
	var offset int
	if enc == EncodingAuto {
		enc, offset = encoding.DetectBOM(data.peek)
	} else {
		_, offset = encoding.DetectBOM(data.peek)
	}

	if cfg.autoSniff {
		delim := sniffDelimiter(data.peek[offset:])
		if delim != 0 {
			cfg.dialect.Delimiter = delim
		}
	}

	combined := io.MultiReader(bytes.NewReader(data.peek[offset:]), data.rest)
	decoded := encoding.NewDecoder(combined, enc)

	it := &RowIterator{
		parser:    parser.New(decoded, cfg.dialect),
		cfg:       cfg,
		streaming: true,
		rowIdx:    -1,
	}

	if cfg.hasHeader && cfg.headerSet {
		row, err := it.parser.Next()
		if err != nil {
			return nil, err
		}
		it.headers = row
	}
	return it, nil
}

func StreamReaderFromFile(path string, opts ...Option) (*RowIterator, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	it, err := StreamReader(file, opts...)
	if err != nil {
		file.Close()
		return nil, err
	}
	it.closer = file
	return it, nil
}

type peekedReader struct {
	peek []byte
	rest io.Reader
}

func peekSample(r io.Reader, n int) (*peekedReader, error) {
	buf := make([]byte, n)
	total := 0
	for total < n {
		m, err := r.Read(buf[total:])
		total += m
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if m == 0 {
			break
		}
	}
	return &peekedReader{peek: buf[:total], rest: r}, nil
}

func (it *RowIterator) Next() bool {
	if it.closed || it.err != nil {
		return false
	}
	if it.streaming {
		row, err := it.parser.Next()
		if err == io.EOF {
			return false
		}
		if err != nil {
			it.err = err
			return false
		}
		it.current = row
		it.rowIdx++
		return true
	}
	it.idx++
	if it.idx >= len(it.rows) {
		return false
	}
	it.current = it.rows[it.idx]
	it.rowIdx = it.idx
	return true
}

func (it *RowIterator) Row() []string {
	cp := make([]string, len(it.current))
	copy(cp, it.current)
	return cp
}

func (it *RowIterator) Record() map[string]string {
	if len(it.headers) == 0 {
		return nil
	}
	out := make(map[string]string, len(it.headers))
	for i, h := range it.headers {
		if i < len(it.current) {
			out[h] = it.current[i]
		} else {
			out[h] = ""
		}
	}
	return out
}

func (it *RowIterator) RowIndex() int {
	return it.rowIdx
}

func (it *RowIterator) Headers() []string {
	out := make([]string, len(it.headers))
	copy(out, it.headers)
	return out
}

func (it *RowIterator) Error() error {
	return it.err
}

func (it *RowIterator) Close() error {
	it.closed = true
	if it.closer != nil {
		err := it.closer.Close()
		it.closer = nil
		return err
	}
	return nil
}
