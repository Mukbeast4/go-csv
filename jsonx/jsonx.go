package jsonx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	gocsv "github.com/mukbeast4/go-csv"
)

type Option func(*config)

type config struct {
	typeInfer    bool
	nullForEmpty bool
	pretty       bool
	hasHeader    bool
	fieldNames   []string
}

func defaultConfig() *config {
	return &config{typeInfer: true, hasHeader: true}
}

func WithTypeInfer(enabled bool) Option    { return func(c *config) { c.typeInfer = enabled } }
func WithNullForEmpty(enabled bool) Option { return func(c *config) { c.nullForEmpty = enabled } }
func WithPretty(enabled bool) Option       { return func(c *config) { c.pretty = enabled } }
func WithHeader(enabled bool) Option       { return func(c *config) { c.hasHeader = enabled } }
func WithFieldNames(names []string) Option {
	return func(c *config) { c.fieldNames = append([]string(nil), names...) }
}

func ToJSON(f *gocsv.File, w io.Writer, opts ...Option) error {
	cfg := applyOpts(opts)
	keys, rows, err := prepareRows(f, cfg)
	if err != nil {
		return err
	}
	out := make([]*orderedMap, len(rows))
	for i, row := range rows {
		out[i] = rowToMap(keys, row, cfg)
	}
	enc := json.NewEncoder(w)
	if cfg.pretty {
		enc.SetIndent("", "  ")
	}
	enc.SetEscapeHTML(false)
	return enc.Encode(out)
}

func ToJSONFile(f *gocsv.File, path string, opts ...Option) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return ToJSON(f, file, opts...)
}

func ToNDJSON(f *gocsv.File, w io.Writer, opts ...Option) error {
	cfg := applyOpts(opts)
	keys, rows, err := prepareRows(f, cfg)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, row := range rows {
		if err := enc.Encode(rowToMap(keys, row, cfg)); err != nil {
			return err
		}
	}
	return nil
}

func ToNDJSONFile(f *gocsv.File, path string, opts ...Option) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return ToNDJSON(f, file, opts...)
}

func FromJSON(r io.Reader, opts ...Option) (*gocsv.File, error) {
	cfg := applyOpts(opts)
	dec := json.NewDecoder(r)
	dec.UseNumber()
	var arr []map[string]any
	if err := dec.Decode(&arr); err != nil {
		return nil, err
	}
	return buildFileFromRecords(arr, cfg), nil
}

func FromJSONFile(path string, opts ...Option) (*gocsv.File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return FromJSON(bytes.NewReader(data), opts...)
}

func FromNDJSON(r io.Reader, opts ...Option) (*gocsv.File, error) {
	cfg := applyOpts(opts)
	dec := json.NewDecoder(r)
	dec.UseNumber()
	var records []map[string]any
	for {
		var rec map[string]any
		if err := dec.Decode(&rec); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		records = append(records, rec)
	}
	return buildFileFromRecords(records, cfg), nil
}

func FromNDJSONFile(path string, opts ...Option) (*gocsv.File, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return FromNDJSON(file, opts...)
}

func applyOpts(opts []Option) *config {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func prepareRows(f *gocsv.File, cfg *config) ([]string, [][]string, error) {
	keys := cfg.fieldNames
	if len(keys) == 0 {
		keys = f.Headers()
	}
	rows, err := f.GetRows()
	if err != nil {
		return nil, nil, err
	}
	return keys, rows, nil
}

func rowToMap(keys []string, row []string, cfg *config) *orderedMap {
	m := newOrderedMap(len(keys))
	for i, k := range keys {
		var raw string
		if i < len(row) {
			raw = row[i]
		}
		if raw == "" && cfg.nullForEmpty {
			m.Set(k, nil)
			continue
		}
		if cfg.typeInfer {
			m.Set(k, inferValue(raw))
		} else {
			m.Set(k, raw)
		}
	}
	return m
}

func inferValue(raw string) any {
	if raw == "" {
		return raw
	}
	if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return n
	}
	if f, err := strconv.ParseFloat(raw, 64); err == nil {
		return f
	}
	lower := strings.ToLower(raw)
	if lower == "true" {
		return true
	}
	if lower == "false" {
		return false
	}
	return raw
}

func buildFileFromRecords(records []map[string]any, cfg *config) *gocsv.File {
	headers := cfg.fieldNames
	if len(headers) == 0 {
		seen := map[string]bool{}
		for _, rec := range records {
			for _, k := range sortedKeys(rec) {
				if !seen[k] {
					seen[k] = true
					headers = append(headers, k)
				}
			}
		}
	}
	out := gocsv.NewFile()
	out.SetHeaders(headers)
	for _, rec := range records {
		row := make([]string, len(headers))
		for i, h := range headers {
			row[i] = anyToString(rec[h])
		}
		out.AppendStrRow(row)
	}
	return out
}

func sortedKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	// stdlib map iteration is randomized; sort for stable first-seen ordering
	// across JSON parses of the same input.
	simpleSort(out)
	return out
}

func simpleSort(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}

func anyToString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case json.Number:
		return t.String()
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case int64:
		return strconv.FormatInt(t, 10)
	case map[string]any, []any:
		b, _ := json.Marshal(v)
		return string(b)
	default:
		return fmt.Sprintf("%v", v)
	}
}
