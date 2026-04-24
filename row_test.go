package gocsv

import "testing"

func TestAppendRow(t *testing.T) {
	f := NewFile()
	f.AppendRow([]any{"a", 1, 3.14, true})
	if f.RowCount() != 1 {
		t.Error("expected 1 row")
	}
	row, _ := f.GetRow(0)
	if len(row) != 4 {
		t.Errorf("fields: %d", len(row))
	}
	if row[0] != "a" || row[1] != "1" || row[2] != "3.14" || row[3] != "true" {
		t.Errorf("row: %v", row)
	}
}

func TestInsertRow(t *testing.T) {
	f := NewFile()
	f.AppendRow([]any{"a"})
	f.AppendRow([]any{"c"})
	f.InsertRow(1, []any{"b"})
	rows, _ := f.GetRows()
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
	if rows[1][0] != "b" {
		t.Errorf("inserted wrong: %v", rows)
	}
}

func TestRemoveRow(t *testing.T) {
	f := NewFile()
	f.AppendRow([]any{"a"})
	f.AppendRow([]any{"b"})
	f.AppendRow([]any{"c"})
	f.RemoveRow(1)
	rows, _ := f.GetRows()
	if len(rows) != 2 || rows[0][0] != "a" || rows[1][0] != "c" {
		t.Errorf("after remove: %v", rows)
	}
}

func TestSetRowValues(t *testing.T) {
	f := NewFile()
	f.SetRowValues(0, []any{"x", "y"})
	f.SetRowValues(2, []any{"z"})
	rows, _ := f.GetRows()
	if len(rows) != 3 {
		t.Fatalf("rows: %d", len(rows))
	}
	if rows[0][0] != "x" || rows[2][0] != "z" {
		t.Errorf("set row: %v", rows)
	}
}

func TestRemoveRowOutOfRange(t *testing.T) {
	f := NewFile()
	if err := f.RemoveRow(0); err == nil {
		t.Error("expected error")
	}
}

func TestAppendStrRow(t *testing.T) {
	f := NewFile()
	f.AppendStrRow([]string{"a", "b"})
	row, _ := f.GetRow(0)
	if row[0] != "a" || row[1] != "b" {
		t.Errorf("row: %v", row)
	}
}

func TestClearRows(t *testing.T) {
	f := NewFile()
	f.AppendRow([]any{"x"})
	f.ClearRows()
	if f.RowCount() != 0 {
		t.Error("expected 0 rows after clear")
	}
}
