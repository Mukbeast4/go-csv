package marshal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"

	gocsv "github.com/mukbeast4/go-csv"
)

func Unmarshal(data []byte, v any) error {
	return UnmarshalFrom(bytes.NewReader(data), v)
}

func UnmarshalFile(path string, v any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return UnmarshalFrom(file, v)
}

func UnmarshalFrom(r io.Reader, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("marshal: target must be non-nil pointer")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Slice {
		return fmt.Errorf("marshal: target must point to slice, got %s", rv.Kind())
	}
	elemType := rv.Type().Elem()
	ptr := false
	if elemType.Kind() == reflect.Pointer {
		ptr = true
		elemType = elemType.Elem()
	}
	info, err := getStructInfo(elemType)
	if err != nil {
		return err
	}
	it, err := gocsv.StreamReader(r, gocsv.WithHeader(true))
	if err != nil {
		return err
	}
	defer it.Close()
	headers := it.Headers()
	colMap := buildColMap(info, headers)
	restCols := buildRestCols(info, colMap, headers)
	slice := reflect.MakeSlice(rv.Type(), 0, 0)
	for it.Next() {
		row := it.Row()
		elem := reflect.New(elemType).Elem()
		if err := rowToStruct(row, elem, info, colMap); err != nil {
			return err
		}
		if info.RestField != nil && len(restCols) > 0 {
			if err := fillRestMap(elem, info, headers, row, restCols); err != nil {
				return err
			}
		}
		if ptr {
			p := reflect.New(elemType)
			p.Elem().Set(elem)
			slice = reflect.Append(slice, p)
		} else {
			slice = reflect.Append(slice, elem)
		}
	}
	if err := it.Error(); err != nil {
		return err
	}
	rv.Set(slice)
	return nil
}

func buildColMap(info *structInfo, headers []string) []int {
	colMap := make([]int, len(info.Fields))
	for i, f := range info.Fields {
		colMap[i] = -1
		for j, h := range headers {
			if h == f.Tag.Name {
				colMap[i] = j
				break
			}
		}
	}
	return colMap
}

func buildRestCols(info *structInfo, colMap []int, headers []string) []int {
	if info.RestField == nil {
		return nil
	}
	matched := map[int]bool{}
	for _, c := range colMap {
		if c >= 0 {
			matched[c] = true
		}
	}
	var out []int
	for i := range headers {
		if !matched[i] {
			out = append(out, i)
		}
	}
	sort.Ints(out)
	return out
}

func fillRestMap(target reflect.Value, info *structInfo, headers []string, row []string, restCols []int) error {
	m := fieldByIndexWrite(target, info.RestField.Index)
	if m.Kind() == reflect.Pointer {
		if m.IsNil() {
			m.Set(reflect.New(m.Type().Elem()))
		}
		m = m.Elem()
	}
	if m.Kind() != reflect.Map {
		return fmt.Errorf("marshal: rest field must be map")
	}
	if m.IsNil() {
		m.Set(reflect.MakeMap(m.Type()))
	}
	for _, ci := range restCols {
		key := headers[ci]
		val := ""
		if ci < len(row) {
			val = row[ci]
		}
		m.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(val))
	}
	return nil
}

func rowToStruct(row []string, target reflect.Value, info *structInfo, colMap []int) error {
	for i, f := range info.Fields {
		col := colMap[i]
		if col < 0 {
			if f.Tag.Required {
				return fmt.Errorf("marshal: missing required column %q", f.Tag.Name)
			}
			continue
		}
		var raw string
		if col < len(row) {
			raw = row[col]
		}
		field := fieldByIndexWrite(target, f.Index)
		if err := decodeValue(raw, field, f.Tag); err != nil {
			return err
		}
	}
	return nil
}

func fieldByIndexWrite(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}
	return v
}

type Decoder struct {
	it       *gocsv.RowIterator
	elem     reflect.Type
	info     *structInfo
	colMap   []int
	restCols []int
	headers  []string
	err      error
	hasNext  bool
}

func NewDecoder(r io.Reader, elemSample any) (*Decoder, error) {
	t := reflect.TypeOf(elemSample)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	info, err := getStructInfo(t)
	if err != nil {
		return nil, err
	}
	it, err := gocsv.StreamReader(r, gocsv.WithHeader(true))
	if err != nil {
		return nil, err
	}
	headers := it.Headers()
	colMap := buildColMap(info, headers)
	return &Decoder{
		it:       it,
		elem:     t,
		info:     info,
		colMap:   colMap,
		restCols: buildRestCols(info, colMap, headers),
		headers:  headers,
	}, nil
}

func (d *Decoder) Decode(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer {
		return fmt.Errorf("marshal: Decode target must be pointer")
	}
	if !d.it.Next() {
		if err := d.it.Error(); err != nil {
			return err
		}
		return io.EOF
	}
	row := d.it.Row()
	target := rv.Elem()
	if err := rowToStruct(row, target, d.info, d.colMap); err != nil {
		return err
	}
	if d.info.RestField != nil && len(d.restCols) > 0 {
		if err := fillRestMap(target, d.info, d.headers, row, d.restCols); err != nil {
			return err
		}
	}
	return nil
}

func (d *Decoder) Close() error {
	return d.it.Close()
}

type Encoder struct {
	sw       *gocsv.StreamWriter
	info     *structInfo
	written  bool
	elem     reflect.Type
	restKeys []string
}

func NewEncoder(w io.Writer, elemSample any) (*Encoder, error) {
	t := reflect.TypeOf(elemSample)
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	info, err := getStructInfo(t)
	if err != nil {
		return nil, err
	}
	return &Encoder{
		sw:   gocsv.NewStreamWriter(w),
		info: info,
		elem: t,
	}, nil
}

func (e *Encoder) SetRestKeys(keys []string) {
	e.restKeys = append([]string(nil), keys...)
	sort.Strings(e.restKeys)
}

func (e *Encoder) Encode(v any) error {
	if !e.written {
		headers := append([]string{}, e.info.Headers...)
		if e.info.RestField != nil {
			if e.restKeys == nil {
				return fmt.Errorf("marshal: struct has rest field; call SetRestKeys before Encode")
			}
			for _, k := range e.restKeys {
				if _, collides := e.info.ByName[k]; collides {
					return fmt.Errorf("marshal: rest key %q collides with struct field", k)
				}
			}
			headers = append(headers, e.restKeys...)
		}
		if err := e.sw.WriteHeader(headers); err != nil {
			return err
		}
		e.written = true
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	row, err := structToRow(rv, e.info)
	if err != nil {
		return err
	}
	if e.info.RestField != nil {
		row = appendRestValues(row, rv, e.info, e.restKeys)
	}
	return e.sw.WriteStrRow(row)
}

func (e *Encoder) EncodeAll(slice any) error {
	rv := reflect.ValueOf(slice)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice {
		return fmt.Errorf("marshal: EncodeAll expects slice, got %s", rv.Kind())
	}
	for i := 0; i < rv.Len(); i++ {
		if err := e.Encode(rv.Index(i).Interface()); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) Close() error {
	return e.sw.Close()
}
