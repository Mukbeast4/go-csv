package marshal

import (
	"encoding"
	"reflect"
	"time"
)

type Marshaler interface {
	MarshalCSV() (string, error)
}

type Unmarshaler interface {
	UnmarshalCSV(s string) error
}

var timeType = reflect.TypeOf(time.Time{})

func isTimeType(t reflect.Type) bool {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t == timeType
}

func tryMarshaler(v reflect.Value) (string, bool, error) {
	if v.Kind() == reflect.Pointer && v.IsNil() {
		return "", true, nil
	}
	if v.IsValid() && v.CanInterface() {
		if m, ok := v.Interface().(Marshaler); ok {
			s, err := m.MarshalCSV()
			return s, true, err
		}
	}
	if v.CanAddr() {
		if m, ok := v.Addr().Interface().(Marshaler); ok {
			s, err := m.MarshalCSV()
			return s, true, err
		}
	}
	if v.IsValid() && v.CanInterface() {
		if m, ok := v.Interface().(encoding.TextMarshaler); ok {
			b, err := m.MarshalText()
			return string(b), true, err
		}
	}
	if v.CanAddr() {
		if m, ok := v.Addr().Interface().(encoding.TextMarshaler); ok {
			b, err := m.MarshalText()
			return string(b), true, err
		}
	}
	return "", false, nil
}

func tryUnmarshaler(target reflect.Value, raw string) (bool, error) {
	if target.Kind() == reflect.Pointer {
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}
		if u, ok := target.Interface().(Unmarshaler); ok {
			return true, u.UnmarshalCSV(raw)
		}
		if u, ok := target.Interface().(encoding.TextUnmarshaler); ok {
			return true, u.UnmarshalText([]byte(raw))
		}
	}
	if target.CanAddr() {
		if u, ok := target.Addr().Interface().(Unmarshaler); ok {
			return true, u.UnmarshalCSV(raw)
		}
		if u, ok := target.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return true, u.UnmarshalText([]byte(raw))
		}
	}
	return false, nil
}
