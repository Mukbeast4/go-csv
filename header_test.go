package gocsv

import "testing"

func TestSetHeaders(t *testing.T) {
	f := NewFile()
	f.SetHeaders([]string{"a", "b"})
	if !f.HasHeader() {
		t.Error("expected header")
	}
	h := f.Headers()
	if len(h) != 2 || h[0] != "a" {
		t.Errorf("headers: %v", h)
	}
}

func TestGetByHeader(t *testing.T) {
	f := NewFile()
	f.SetHeaders([]string{"name", "age"})
	f.AppendRow([]any{"Alice", 30})
	v, err := f.GetByHeader(0, "age")
	if err != nil {
		t.Fatal(err)
	}
	if v != "30" {
		t.Errorf("got %q", v)
	}
}

func TestGetByHeaderNoHeader(t *testing.T) {
	f := NewFile()
	_, err := f.GetByHeader(0, "x")
	if err != ErrNoHeader {
		t.Errorf("expected ErrNoHeader, got %v", err)
	}
}

func TestSetByHeader(t *testing.T) {
	f := NewFile()
	f.SetHeaders([]string{"name", "age"})
	f.AppendRow([]any{"Alice", 30})
	f.SetByHeader(0, "age", 31)
	v, _ := f.GetByHeader(0, "age")
	if v != "31" {
		t.Errorf("got %q", v)
	}
}

func TestAppendRecord(t *testing.T) {
	f := NewFile()
	f.SetHeaders([]string{"name", "age", "city"})
	f.AppendRecord(map[string]any{"name": "Alice", "age": 30, "city": "Paris"})
	r, err := f.GetRecord(0)
	if err != nil {
		t.Fatal(err)
	}
	if r["name"] != "Alice" || r["age"] != "30" || r["city"] != "Paris" {
		t.Errorf("record: %v", r)
	}
}

func TestGetRecords(t *testing.T) {
	f := NewFile()
	f.SetHeaders([]string{"id", "v"})
	f.AppendRecord(map[string]any{"id": 1, "v": "a"})
	f.AppendRecord(map[string]any{"id": 2, "v": "b"})
	recs, err := f.GetRecords()
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 2 {
		t.Fatalf("count: %d", len(recs))
	}
	if recs[0]["v"] != "a" || recs[1]["v"] != "b" {
		t.Errorf("records: %v", recs)
	}
}

func TestHeaderIndex(t *testing.T) {
	f := NewFile()
	f.SetHeaders([]string{"a", "b", "c"})
	if _, ok := f.HeaderIndex("b"); !ok {
		t.Error("should find b")
	}
	if _, ok := f.HeaderIndex("missing"); ok {
		t.Error("should not find missing")
	}
}
