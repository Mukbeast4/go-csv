package gocsv

import (
	"errors"
	"fmt"
)

var (
	ErrFileClosed       = errors.New("gocsv: file is closed")
	ErrInvalidCell      = errors.New("gocsv: invalid cell reference")
	ErrInvalidCoords    = errors.New("gocsv: invalid coordinates")
	ErrInvalidRange     = errors.New("gocsv: invalid range")
	ErrRowOutOfRange    = errors.New("gocsv: row out of range")
	ErrColumnOutOfRange = errors.New("gocsv: column out of range")
	ErrHeaderNotFound   = errors.New("gocsv: header not found")
	ErrNoHeader         = errors.New("gocsv: file has no header row")
	ErrFieldCount       = errors.New("gocsv: row field count mismatch")
	ErrUnsupportedType  = errors.New("gocsv: unsupported value type")
	ErrBareQuote        = errors.New("gocsv: bare quote in unquoted field")
	ErrUnclosedQuote    = errors.New("gocsv: unclosed quoted field")
	ErrEncodingInvalid  = errors.New("gocsv: invalid encoding")
	ErrInvalidDelimiter = errors.New("gocsv: invalid delimiter")
	ErrStreamClosed     = errors.New("gocsv: stream is closed")
)

type ParseError struct {
	Line   int
	Column int
	Row    int
	Offset int64
	Err    error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("gocsv: parse error at line %d column %d (offset %d): %v", e.Line, e.Column, e.Offset, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

type CellError struct {
	Cell string
	Row  int
	Col  int
	Err  error
}

func (e *CellError) Error() string {
	if e.Cell != "" {
		return fmt.Sprintf("gocsv: cell %s: %v", e.Cell, e.Err)
	}
	return fmt.Sprintf("gocsv: row %d col %d: %v", e.Row, e.Col, e.Err)
}

func (e *CellError) Unwrap() error {
	return e.Err
}

func cellErr(cell string, row, col int, err error) *CellError {
	return &CellError{Cell: cell, Row: row, Col: col, Err: err}
}
