package marshal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"

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
	slice := reflect.MakeSlice(rv.Type(), 0, 0)
	for it.Next() {
		row := it.Row()
		elem := reflect.New(elemType).Elem()
		if err := rowToStruct(row, elem, info, colMap); err != nil {
			return err
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
	it      *gocsv.RowIterator
	elem    reflect.Type
	info    *structInfo
	colMap  []int
	err     error
	hasNext bool
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
	return &Decoder{
		it:     it,
		elem:   t,
		info:   info,
		colMap: buildColMap(info, it.Headers()),
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
	return rowToStruct(d.it.Row(), rv.Elem(), d.info, d.colMap)
}

func (d *Decoder) Close() error {
	return d.it.Close()
}

type Encoder struct {
	sw      *gocsv.StreamWriter
	info    *structInfo
	written bool
	elem    reflect.Type
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

func (e *Encoder) Encode(v any) error {
	if !e.written {
		if err := e.sw.WriteHeader(e.info.Headers); err != nil {
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
