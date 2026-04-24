package jsonx

import (
	"encoding/json"
	"io"

	gocsv "github.com/mukbeast4/go-csv"
)

type Encoder struct {
	w       io.Writer
	enc     *json.Encoder
	headers []string
	cfg     *config
}

func NewEncoder(w io.Writer, headers []string, opts ...Option) *Encoder {
	cfg := applyOpts(opts)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if cfg.pretty {
		enc.SetIndent("", "  ")
	}
	hdrs := append([]string(nil), headers...)
	return &Encoder{w: w, enc: enc, headers: hdrs, cfg: cfg}
}

func (e *Encoder) Encode(row []string) error {
	m := rowToMap(e.headers, row, e.cfg)
	return e.enc.Encode(m)
}

func (e *Encoder) EncodeRecord(record map[string]string) error {
	row := make([]string, len(e.headers))
	for i, h := range e.headers {
		row[i] = record[h]
	}
	return e.Encode(row)
}

type Decoder struct {
	dec     *json.Decoder
	headers []string
	cfg     *config
	current []string
	err     error
	rowIdx  int
}

func NewDecoder(r io.Reader, opts ...Option) (*Decoder, error) {
	cfg := applyOpts(opts)
	dec := json.NewDecoder(r)
	dec.UseNumber()
	d := &Decoder{dec: dec, cfg: cfg, rowIdx: -1}
	if err := d.primeHeaders(); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *Decoder) primeHeaders() error {
	if len(d.cfg.fieldNames) > 0 {
		d.headers = append([]string(nil), d.cfg.fieldNames...)
		return nil
	}
	// Peek at the first record to discover headers, then consume it.
	var first map[string]any
	if err := d.dec.Decode(&first); err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	d.headers = sortedKeys(first)
	d.current = recordToRow(d.headers, first)
	return nil
}

func (d *Decoder) Headers() []string {
	out := make([]string, len(d.headers))
	copy(out, d.headers)
	return out
}

func (d *Decoder) Next() bool {
	if d.err != nil {
		return false
	}
	if d.current != nil {
		return true
	}
	var rec map[string]any
	err := d.dec.Decode(&rec)
	if err == io.EOF {
		return false
	}
	if err != nil {
		d.err = err
		return false
	}
	d.rowIdx++
	d.current = recordToRow(d.headers, rec)
	return true
}

func (d *Decoder) Row() []string {
	out := make([]string, len(d.current))
	copy(out, d.current)
	d.current = nil
	if d.rowIdx < 0 {
		d.rowIdx = 0
	}
	return out
}

func (d *Decoder) Record() map[string]string {
	if len(d.headers) == 0 {
		return nil
	}
	out := make(map[string]string, len(d.headers))
	for i, h := range d.headers {
		if i < len(d.current) {
			out[h] = d.current[i]
		}
	}
	d.current = nil
	return out
}

func (d *Decoder) Error() error { return d.err }

func (d *Decoder) Close() error { return nil }

func recordToRow(headers []string, rec map[string]any) []string {
	row := make([]string, len(headers))
	for i, h := range headers {
		row[i] = anyToString(rec[h])
	}
	return row
}

// StreamReader returns a *gocsv.RowIterator backed by an NDJSON stream.
func StreamReader(r io.Reader, opts ...Option) (*gocsv.RowIterator, error) {
	d, err := NewDecoder(r, opts...)
	if err != nil {
		return nil, err
	}
	return gocsv.NewRowIteratorFromFunc(d.headers, func() ([]string, error) {
		if !d.Next() {
			if d.err != nil {
				return nil, d.err
			}
			return nil, io.EOF
		}
		return d.Row(), nil
	}, nil), nil
}
