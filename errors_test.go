package gocsv

import (
	"errors"
	"testing"
)

func TestParseErrorUnwrap(t *testing.T) {
	inner := errors.New("inner")
	pe := &ParseError{Line: 5, Column: 3, Offset: 42, Err: inner}
	if !errors.Is(pe, inner) {
		t.Error("ParseError should unwrap to inner")
	}
	if pe.Error() == "" {
		t.Error("empty message")
	}
}

func TestCellErrorUnwrap(t *testing.T) {
	inner := ErrInvalidCell
	ce := &CellError{Cell: "A1", Err: inner}
	if !errors.Is(ce, ErrInvalidCell) {
		t.Error("CellError should unwrap to ErrInvalidCell")
	}
}

func TestCellErrorMessages(t *testing.T) {
	ce := &CellError{Cell: "A1", Err: ErrInvalidCell}
	if ce.Error() == "" {
		t.Error("empty message with Cell")
	}
	ce2 := &CellError{Row: 2, Col: 3, Err: ErrInvalidCell}
	if ce2.Error() == "" {
		t.Error("empty message without Cell")
	}
}
