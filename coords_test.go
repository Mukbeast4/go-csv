package gocsv

import (
	"errors"
	"testing"
)

func TestCellNameToCoordinates(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		col     int
		row     int
		wantErr bool
	}{
		{"A1", "A1", 1, 1, false},
		{"Z26", "Z26", 26, 26, false},
		{"AA27", "AA27", 27, 27, false},
		{"AB1", "AB1", 28, 1, false},
		{"ZZ1", "ZZ1", 702, 1, false},
		{"AAA1", "AAA1", 703, 1, false},
		{"lowercase", "a1", 1, 1, false},
		{"empty", "", 0, 0, true},
		{"only letters", "ABC", 0, 0, true},
		{"only digits", "123", 0, 0, true},
		{"invalid chars", "A!1", 0, 0, true},
		{"leading space", "  A1", 1, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col, row, err := CellNameToCoordinates(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.in)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if col != tt.col || row != tt.row {
				t.Fatalf("got (%d,%d), want (%d,%d)", col, row, tt.col, tt.row)
			}
		})
	}
}

func TestCoordinatesToCellName(t *testing.T) {
	tests := []struct {
		col, row int
		want     string
		wantErr  bool
	}{
		{1, 1, "A1", false},
		{26, 1, "Z1", false},
		{27, 1, "AA1", false},
		{702, 100, "ZZ100", false},
		{703, 1, "AAA1", false},
		{0, 1, "", true},
		{1, 0, "", true},
	}
	for _, tt := range tests {
		got, err := CoordinatesToCellName(tt.col, tt.row)
		if tt.wantErr {
			if err == nil {
				t.Errorf("CoordinatesToCellName(%d,%d): expected error", tt.col, tt.row)
			}
			continue
		}
		if err != nil {
			t.Errorf("CoordinatesToCellName(%d,%d): %v", tt.col, tt.row, err)
			continue
		}
		if got != tt.want {
			t.Errorf("got %q, want %q", got, tt.want)
		}
	}
}

func TestRoundTripCoords(t *testing.T) {
	for col := 1; col <= 1000; col++ {
		for _, row := range []int{1, 10, 100, 1000} {
			name, err := CoordinatesToCellName(col, row)
			if err != nil {
				t.Fatalf("CoordinatesToCellName(%d,%d): %v", col, row, err)
			}
			c, r, err := CellNameToCoordinates(name)
			if err != nil {
				t.Fatalf("CellNameToCoordinates(%q): %v", name, err)
			}
			if c != col || r != row {
				t.Fatalf("round-trip mismatch: (%d,%d) -> %q -> (%d,%d)", col, row, name, c, r)
			}
		}
	}
}

func TestSplitCellRange(t *testing.T) {
	sc, sr, ec, er, err := splitCellRange("A1:C10")
	if err != nil {
		t.Fatal(err)
	}
	if sc != 1 || sr != 1 || ec != 3 || er != 10 {
		t.Errorf("got (%d,%d,%d,%d), want (1,1,3,10)", sc, sr, ec, er)
	}
	sc, sr, ec, er, err = splitCellRange("B5")
	if err != nil {
		t.Fatal(err)
	}
	if sc != 2 || sr != 5 || ec != 2 || er != 5 {
		t.Errorf("single cell: got (%d,%d,%d,%d)", sc, sr, ec, er)
	}
}

func TestCellNameToCoordinatesErrorIs(t *testing.T) {
	_, _, err := CellNameToCoordinates("")
	if !errors.Is(err, ErrInvalidCell) {
		t.Errorf("expected ErrInvalidCell, got %v", err)
	}
}
