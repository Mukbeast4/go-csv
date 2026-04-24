package jsonx

import (
	"bytes"
	"encoding/json"
)

type orderedMap struct {
	keys   []string
	values map[string]any
}

func newOrderedMap(cap int) *orderedMap {
	return &orderedMap{
		keys:   make([]string, 0, cap),
		values: make(map[string]any, cap),
	}
}

func (m *orderedMap) Set(key string, value any) {
	if _, exists := m.values[key]; !exists {
		m.keys = append(m.keys, key)
	}
	m.values[key] = value
}

func (m *orderedMap) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, k := range m.keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		kb, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		buf.Write(kb)
		buf.WriteByte(':')
		vb, err := json.Marshal(m.values[k])
		if err != nil {
			return nil, err
		}
		buf.Write(vb)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}
