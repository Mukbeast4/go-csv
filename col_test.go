package gocsv

import "testing"

func TestGetCol(t *testing.T) {
	f := NewFile()
	f.AppendRow([]any{"a", 1, "x"})
	f.AppendRow([]any{"b", 2, "y"})
	f.AppendRow([]any{"c", 3, "z"})
	col, err := f.GetCol(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(col) != 3 || col[0] != "1" || col[2] != "3" {
		t.Errorf("col: %v", col)
	}
}

func TestGetColByName(t *testing.T) {
	f := NewFile()
	f.SetHeaders([]string{"id", "name"})
	f.AppendRow([]any{"1", "Alice"})
	f.AppendRow([]any{"2", "Bob"})
	col, err := f.GetColByName("name")
	if err != nil {
		t.Fatal(err)
	}
	if col[0] != "Alice" || col[1] != "Bob" {
		t.Errorf("col: %v", col)
	}
}

func TestSetColValues(t *testing.T) {
	f := NewFile()
	f.AppendRow([]any{"", ""})
	f.AppendRow([]any{"", ""})
	f.SetColValues(1, []any{"new1", "new2"})
	rows, _ := f.GetRows()
	if rows[0][1] != "new1" || rows[1][1] != "new2" {
		t.Errorf("rows: %v", rows)
	}
}

func TestInsertCol(t *testing.T) {
	f := NewFile()
	f.SetHeaders([]string{"a", "c"})
	f.AppendRow([]any{"1", "3"})
	f.AppendRow([]any{"4", "6"})
	f.InsertCol(1, "b", []any{"2", "5"})
	headers := f.Headers()
	if len(headers) != 3 || headers[1] != "b" {
		t.Errorf("headers: %v", headers)
	}
	rows, _ := f.GetRows()
	if rows[0][1] != "2" || rows[1][1] != "5" {
		t.Errorf("rows: %v", rows)
	}
}

func TestRemoveCol(t *testing.T) {
	f := NewFile()
	f.SetHeaders([]string{"a", "b", "c"})
	f.AppendRow([]any{"1", "2", "3"})
	f.RemoveCol(1)
	headers := f.Headers()
	if len(headers) != 2 {
		t.Errorf("headers: %v", headers)
	}
	rows, _ := f.GetRows()
	if rows[0][0] != "1" || rows[0][1] != "3" {
		t.Errorf("row: %v", rows[0])
	}
}

func TestRemoveColByName(t *testing.T) {
	f := NewFile()
	f.SetHeaders([]string{"a", "b"})
	f.AppendRow([]any{"1", "2"})
	if err := f.RemoveColByName("missing"); err != ErrHeaderNotFound {
		t.Errorf("expected ErrHeaderNotFound, got %v", err)
	}
	if err := f.RemoveColByName("a"); err != nil {
		t.Fatal(err)
	}
	if f.ColCount() != 1 {
		t.Error("expected 1 col after remove")
	}
}
