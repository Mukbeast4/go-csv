package marshal

import (
	"fmt"
	"reflect"
	"sync"
)

type structFieldInfo struct {
	Index []int
	Tag   fieldTag
	Type  reflect.Type
}

type structInfo struct {
	Fields    []structFieldInfo
	ByName    map[string]int
	Headers   []string
	RestField *structFieldInfo
}

var structCache sync.Map

func getStructInfo(t reflect.Type) (*structInfo, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("marshal: expected struct, got %s", t.Kind())
	}
	if cached, ok := structCache.Load(t); ok {
		return cached.(*structInfo), nil
	}
	info := &structInfo{ByName: make(map[string]int)}
	if err := collectFields(t, nil, "", info, map[reflect.Type]bool{}); err != nil {
		return nil, err
	}
	structCache.Store(t, info)
	return info, nil
}

func collectFields(t reflect.Type, index []int, prefix string, info *structInfo, visited map[reflect.Type]bool) error {
	if visited[t] {
		return fmt.Errorf("marshal: cycle detected on %s via flatten", t)
	}
	visited[t] = true
	defer delete(visited, t)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() && !f.Anonymous {
			continue
		}
		fieldIndex := append(append([]int{}, index...), i)

		if f.Anonymous && f.Type.Kind() == reflect.Struct && !isTimeType(f.Type) {
			if err := collectFields(f.Type, fieldIndex, prefix, info, visited); err != nil {
				return err
			}
			continue
		}

		tag := parseTag(f.Tag, f.Name)
		if tag.Skip {
			continue
		}

		if tag.Rest {
			if info.RestField != nil {
				return fmt.Errorf("marshal: struct %s has multiple rest fields", t)
			}
			mt := f.Type
			for mt.Kind() == reflect.Pointer {
				mt = mt.Elem()
			}
			if mt.Kind() != reflect.Map || mt.Key().Kind() != reflect.String || mt.Elem().Kind() != reflect.String {
				return fmt.Errorf("marshal: rest field %s must be map[string]string, got %s", f.Name, f.Type)
			}
			info.RestField = &structFieldInfo{
				Index: fieldIndex,
				Tag:   tag,
				Type:  f.Type,
			}
			continue
		}

		if tag.Flatten {
			ft := f.Type
			for ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			if ft.Kind() != reflect.Struct || isTimeType(ft) {
				return fmt.Errorf("marshal: flatten requires struct type on field %s (got %s)", f.Name, f.Type)
			}
			sub := prefix
			if tag.Prefix != "" {
				sub = sub + tag.Prefix
			} else {
				sub = sub + tag.Name + "."
			}
			if err := collectFields(ft, fieldIndex, sub, info, visited); err != nil {
				return err
			}
			continue
		}

		name := prefix + tag.Name
		info.ByName[name] = len(info.Fields)
		info.Headers = append(info.Headers, name)
		fi := structFieldInfo{
			Index: fieldIndex,
			Tag:   tag,
			Type:  f.Type,
		}
		fi.Tag.Name = name
		info.Fields = append(info.Fields, fi)
	}
	return nil
}
