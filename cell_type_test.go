package gocsv

import (
	"testing"
	"time"
)

func TestDetectCellType(t *testing.T) {
	tests := []struct {
		name string
		in   any
		ct   CellType
		str  string
	}{
		{"nil", nil, CellTypeEmpty, ""},
		{"empty string", "", CellTypeEmpty, ""},
		{"string", "hello", CellTypeString, "hello"},
		{"int", 42, CellTypeInt, "42"},
		{"int64", int64(100), CellTypeInt, "100"},
		{"float", 3.14, CellTypeFloat, "3.14"},
		{"bool true", true, CellTypeBool, "true"},
		{"bool false", false, CellTypeBool, "false"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct, s := detectCellType(tt.in)
			if ct != tt.ct || s != tt.str {
				t.Errorf("got (%v,%q), want (%v,%q)", ct, s, tt.ct, tt.str)
			}
		})
	}
}

func TestDetectCellTypeTime(t *testing.T) {
	now := time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC)
	ct, s := detectCellType(now)
	if ct != CellTypeDate {
		t.Errorf("expected CellTypeDate, got %v", ct)
	}
	if s == "" {
		t.Error("expected non-empty date string")
	}
}

func TestInferCellType(t *testing.T) {
	tests := []struct {
		in string
		ct CellType
	}{
		{"", CellTypeEmpty},
		{"hello", CellTypeString},
		{"42", CellTypeInt},
		{"3.14", CellTypeFloat},
		{"true", CellTypeBool},
		{"TRUE", CellTypeBool},
		{"2026-04-24", CellTypeDate},
	}
	for _, tt := range tests {
		got := inferCellType(tt.in)
		if got != tt.ct {
			t.Errorf("inferCellType(%q): got %v, want %v", tt.in, got, tt.ct)
		}
	}
}

func TestParseInt(t *testing.T) {
	n, err := parseInt("42")
	if err != nil || n != 42 {
		t.Errorf("parseInt(42): got %d, %v", n, err)
	}
	n, err = parseInt("3.9")
	if err != nil || n != 3 {
		t.Errorf("parseInt(3.9): got %d, %v", n, err)
	}
	_, err = parseInt("abc")
	if err == nil {
		t.Error("expected error for non-numeric")
	}
}

func TestCellTypeString(t *testing.T) {
	for _, ct := range []CellType{CellTypeEmpty, CellTypeString, CellTypeInt, CellTypeFloat, CellTypeBool, CellTypeDate} {
		if ct.String() == "unknown" {
			t.Errorf("%v should have a name", ct)
		}
	}
}
