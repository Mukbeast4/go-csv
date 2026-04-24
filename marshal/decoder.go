package marshal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func decodeValue(raw string, target reflect.Value, tag fieldTag) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if tag.Required {
			return fmt.Errorf("marshal: required field %q is empty", tag.Name)
		}
		if tag.OmitEmpty {
			return nil
		}
	}
	if tag.JSON {
		if raw == "" {
			return nil
		}
		if target.Kind() == reflect.Pointer {
			if target.IsNil() {
				target.Set(reflect.New(target.Type().Elem()))
			}
			return json.Unmarshal([]byte(raw), target.Interface())
		}
		if !target.CanAddr() {
			return fmt.Errorf("marshal: json tag requires addressable field")
		}
		return json.Unmarshal([]byte(raw), target.Addr().Interface())
	}
	if !isTimeType(target.Type()) {
		if ok, err := tryUnmarshaler(target, raw); ok {
			return err
		}
	}
	if target.Kind() == reflect.Pointer {
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}
		target = target.Elem()
	}
	switch target.Kind() {
	case reflect.String:
		target.SetString(raw)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if raw == "" {
			return nil
		}
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return fmt.Errorf("marshal: %q as int: %w", raw, err)
		}
		target.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if raw == "" {
			return nil
		}
		n, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return fmt.Errorf("marshal: %q as uint: %w", raw, err)
		}
		target.SetUint(n)
	case reflect.Float32, reflect.Float64:
		if raw == "" {
			return nil
		}
		n, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("marshal: %q as float: %w", raw, err)
		}
		target.SetFloat(n)
	case reflect.Bool:
		if raw == "" {
			return nil
		}
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("marshal: %q as bool: %w", raw, err)
		}
		target.SetBool(b)
	case reflect.Slice:
		return decodeSlice(raw, target, tag)
	case reflect.Struct:
		if target.Type() == reflect.TypeOf(time.Time{}) {
			return decodeTime(raw, target, tag)
		}
		return fmt.Errorf("marshal: unsupported struct type %s", target.Type())
	default:
		return fmt.Errorf("marshal: unsupported kind %s", target.Kind())
	}
	return nil
}

func decodeSlice(raw string, target reflect.Value, tag fieldTag) error {
	sep := tag.Separator
	if sep == "" {
		sep = ","
	}
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, sep)
	slice := reflect.MakeSlice(target.Type(), len(parts), len(parts))
	for i, p := range parts {
		if err := decodeValue(strings.TrimSpace(p), slice.Index(i), tag); err != nil {
			return err
		}
	}
	target.Set(slice)
	return nil
}

func decodeTime(raw string, target reflect.Value, tag fieldTag) error {
	if raw == "" {
		return nil
	}
	formats := []string{tag.Format, time.RFC3339, "2006-01-02T15:04:05", "2006-01-02", "02/01/2006"}
	for _, f := range formats {
		if f == "" {
			continue
		}
		t, err := time.Parse(f, raw)
		if err == nil {
			target.Set(reflect.ValueOf(t))
			return nil
		}
	}
	return fmt.Errorf("marshal: cannot parse %q as time", raw)
}
