package gocsv

import (
	"testing"
	"time"
)

func TestCellSetGetStr(t *testing.T) {
	f := NewFile()
	if err := f.SetCellStr("A1", "hello"); err != nil {
		t.Fatal(err)
	}
	v, err := f.GetCellStr("A1")
	if err != nil {
		t.Fatal(err)
	}
	if v != "hello" {
		t.Errorf("got %q", v)
	}
}

func TestCellSetGetInt(t *testing.T) {
	f := NewFile()
	f.SetCellInt("B2", 42)
	n, err := f.GetCellInt("B2")
	if err != nil {
		t.Fatal(err)
	}
	if n != 42 {
		t.Errorf("got %d", n)
	}
}

func TestCellSetGetFloat(t *testing.T) {
	f := NewFile()
	f.SetCellFloat("C3", 3.14)
	v, err := f.GetCellFloat("C3")
	if err != nil {
		t.Fatal(err)
	}
	if v != 3.14 {
		t.Errorf("got %v", v)
	}
}

func TestCellSetGetBool(t *testing.T) {
	f := NewFile()
	f.SetCellBool("A1", true)
	v, err := f.GetCellBool("A1")
	if err != nil {
		t.Fatal(err)
	}
	if !v {
		t.Error("expected true")
	}
}

func TestCellSetGetDate(t *testing.T) {
	f := NewFile()
	now := time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
	f.SetCellDate("A1", now)
	v, err := f.GetCellDate("A1")
	if err != nil {
		t.Fatal(err)
	}
	if !v.Equal(now) {
		t.Errorf("got %v, want %v", v, now)
	}
}

func TestCellGetType(t *testing.T) {
	f := NewFile()
	f.SetCellInt("A1", 42)
	ct, _ := f.GetCellType("A1")
	if ct != CellTypeInt {
		t.Errorf("A1 type: got %v", ct)
	}
	f.SetCellStr("B1", "hello")
	ct, _ = f.GetCellType("B1")
	if ct != CellTypeString {
		t.Errorf("B1 type: got %v", ct)
	}
}

func TestCellInvalidRef(t *testing.T) {
	f := NewFile()
	if err := f.SetCellValue("XYZ", "value"); err == nil {
		t.Error("expected error for invalid cell")
	}
}

func TestCellEmpty(t *testing.T) {
	f := NewFile()
	v, err := f.GetCellStr("A1")
	if err != nil {
		t.Fatal(err)
	}
	if v != "" {
		t.Errorf("expected empty, got %q", v)
	}
}
