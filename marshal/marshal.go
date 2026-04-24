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

	var restKeys []string
	if info.RestField != nil {
		restKeys, err = collectRestKeys(rv, info)
		if err != nil {
			return err
		}
	}

	headers := append([]string{}, info.Headers...)
	headers = append(headers, restKeys...)

	sw := gocsv.NewStreamWriter(w)
	defer sw.Close()
	if err := sw.WriteHeader(headers); err != nil {
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
		if info.RestField != nil {
			row = appendRestValues(row, elem, info, restKeys)
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

func collectRestKeys(rv reflect.Value, info *structInfo) ([]string, error) {
	keys := map[string]struct{}{}
	for i := 0; i < rv.Len(); i++ {
		elem := rv.Index(i)
		for elem.Kind() == reflect.Pointer {
			if elem.IsNil() {
				elem = reflect.Value{}
				break
			}
			elem = elem.Elem()
		}
		if !elem.IsValid() {
			continue
		}
		m := fieldByIndex(elem, info.RestField.Index)
		for m.Kind() == reflect.Pointer {
			if m.IsNil() {
				m = reflect.Value{}
				break
			}
			m = m.Elem()
		}
		if !m.IsValid() || m.Kind() != reflect.Map {
			continue
		}
		for _, k := range m.MapKeys() {
			keys[k.String()] = struct{}{}
		}
	}
	out := make([]string, 0, len(keys))
	for k := range keys {
		if _, collides := info.ByName[k]; collides {
			return nil, fmt.Errorf("marshal: rest map key %q collides with struct field", k)
		}
		out = append(out, k)
	}
	sort.Strings(out)
	return out, nil
}

func appendRestValues(row []string, elem reflect.Value, info *structInfo, keys []string) []string {
	m := fieldByIndex(elem, info.RestField.Index)
	for m.Kind() == reflect.Pointer {
		if m.IsNil() {
			m = reflect.Value{}
			break
		}
		m = m.Elem()
	}
	for _, k := range keys {
		if !m.IsValid() || m.Kind() != reflect.Map {
			row = append(row, "")
			continue
		}
		val := m.MapIndex(reflect.ValueOf(k))
		if !val.IsValid() {
			row = append(row, "")
		} else {
			row = append(row, val.String())
		}
	}
	return row
}
