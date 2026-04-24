package marshal

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"

	gocsv "github.com/mukbeast4/go-csv"
)

func Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := MarshalTo(&buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func MarshalTo(w io.Writer, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice {
		return fmt.Errorf("marshal: expected slice, got %s", rv.Kind())
	}
	elemType := rv.Type().Elem()
	for elemType.Kind() == reflect.Pointer {
		elemType = elemType.Elem()
	}
	info, err := getStructInfo(elemType)
	if err != nil {
		return err
	}
	sw := gocsv.NewStreamWriter(w)
	defer sw.Close()
	if err := sw.WriteHeader(info.Headers); err != nil {
		return err
	}
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)
		for elem.Kind() == reflect.Pointer {
			if elem.IsNil() {
				continue
			}
			elem = elem.Elem()
		}
		row, err := structToRow(elem, info)
		if err != nil {
			return err
		}
		if err := sw.WriteStrRow(row); err != nil {
			return err
		}
	}
	return nil
}

func MarshalFile(path string, v any) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return MarshalTo(file, v)
}

func structToRow(v reflect.Value, info *structInfo) ([]string, error) {
	row := make([]string, len(info.Fields))
	for i, f := range info.Fields {
		field := fieldByIndex(v, f.Index)
		s, err := encodeValue(field, f.Tag)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", f.Tag.Name, err)
		}
		row[i] = s
	}
	return row, nil
}

func fieldByIndex(v reflect.Value, index []int) reflect.Value {
	for _, i := range index {
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				return reflect.Value{}
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}
	return v
}
