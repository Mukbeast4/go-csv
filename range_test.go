package gocsv

import "testing"

func TestRangeValues(t *testing.T) {
	f := NewFile()
	f.AppendRow([]any{"a", "b", "c"})
	f.AppendRow([]any{"1", "2", "3"})
	f.AppendRow([]any{"x", "y", "z"})
	r := f.Range("A1:B2")
	if r.Err() != nil {
		t.Fatal(r.Err())
	}
	vals := r.Values()
	if len(vals) != 2 || len(vals[0]) != 2 {
		t.Fatalf("dim: %v", vals)
	}
	if vals[0][0] != "a" || vals[1][1] != "2" {
		t.Errorf("vals: %v", vals)
	}
}

func TestRangeSetValue(t *testing.T) {
	f := NewFile()
	f.AppendRow([]any{"", "", ""})
	f.AppendRow([]any{"", "", ""})
	r := f.Range("A1:B2").SetValue("X")
	if r.Err() != nil {
		t.Fatal(r.Err())
	}
	rows, _ := f.GetRows()
	if rows[0][0] != "X" || rows[1][1] != "X" {
		t.Errorf("set: %v", rows)
	}
}

func TestRangeSetValues(t *testing.T) {
	f := NewFile()
	r := f.Range("A1:B2").SetValues([][]any{{"a", "b"}, {"c", "d"}})
	if r.Err() != nil {
		t.Fatal(r.Err())
	}
	rows, _ := f.GetRows()
	if rows[0][0] != "a" || rows[1][1] != "d" {
		t.Errorf("values: %v", rows)
	}
}

func TestRangeForEach(t *testing.T) {
	f := NewFile()
	f.AppendRow([]any{"1", "2"})
	f.AppendRow([]any{"3", "4"})
	count := 0
	f.Range("A1:B2").ForEach(func(col, row int, v string) {
		count++
	})
	if count != 4 {
		t.Errorf("count: %d", count)
	}
}

func TestRangeInvalid(t *testing.T) {
	f := NewFile()
	r := f.Range("invalid")
	if r.Err() == nil {
		t.Error("expected error")
	}
}

func TestRangeDimensions(t *testing.T) {
	f := NewFile()
	rows, cols := f.Range("A1:C5").Dimensions()
	if rows != 5 || cols != 3 {
		t.Errorf("dim: %dx%d", rows, cols)
	}
}
