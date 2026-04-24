package marshal

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func encodeValue(v reflect.Value, tag fieldTag) (string, error) {
	if !v.IsValid() {
		return "", nil
	}
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return "", nil
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		return v.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'f', -1, 32), nil
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.Slice:
		return encodeSlice(v, tag)
	case reflect.Struct:
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return encodeTime(v.Interface().(time.Time), tag), nil
		}
		return "", fmt.Errorf("marshal: unsupported struct type %s", v.Type())
	default:
		return fmt.Sprintf("%v", v.Interface()), nil
	}
}

func encodeSlice(v reflect.Value, tag fieldTag) (string, error) {
	sep := tag.Separator
	if sep == "" {
		sep = ","
	}
	parts := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		s, err := encodeValue(v.Index(i), tag)
		if err != nil {
			return "", err
		}
		parts[i] = s
	}
	return strings.Join(parts, sep), nil
}

func encodeTime(t time.Time, tag fieldTag) string {
	format := tag.Format
	if format == "" {
		format = time.RFC3339
	}
	if t.IsZero() {
		return ""
	}
	return t.Format(format)
}
