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
	Fields  []structFieldInfo
	ByName  map[string]int
	Headers []string
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
	collectFields(t, nil, info)
	structCache.Store(t, info)
	return info, nil
}

func collectFields(t reflect.Type, index []int, info *structInfo) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() && !f.Anonymous {
			continue
		}
		fieldIndex := append(append([]int{}, index...), i)
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			collectFields(f.Type, fieldIndex, info)
			continue
		}
		tag := parseTag(f.Tag, f.Name)
		if tag.Skip {
			continue
		}
		info.ByName[tag.Name] = len(info.Fields)
		info.Headers = append(info.Headers, tag.Name)
		info.Fields = append(info.Fields, structFieldInfo{
			Index: fieldIndex,
			Tag:   tag,
			Type:  f.Type,
		})
	}
}
