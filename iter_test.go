package gocsv

import (
	"errors"
	"io"
	"strings"
	"testing"
)

func TestRowIterator(t *testing.T) {
	f := NewFile()
	f.AppendRow([]any{"a", "1"})
	f.AppendRow([]any{"b", "2"})
	f.AppendRow([]any{"c", "3"})
	it, err := f.NewRowIterator()
	if err != nil {
		t.Fatal(err)
	}
	defer it.Close()
	count := 0
	for it.Next() {
		count++
	}
	if it.Error() != nil {
		t.Error(it.Error())
	}
	if count != 3 {
		t.Errorf("count: %d", count)
	}
}

func TestStreamReader(t *testing.T) {
	data := "name,age\nAlice,30\nBob,25\n"
	it, err := StreamReader(strings.NewReader(data), WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	defer it.Close()
	count := 0
	for it.Next() {
		count++
		rec := it.Record()
		if rec == nil {
			t.Error("expected record")
		}
	}
	if count != 2 {
		t.Errorf("count: %d", count)
	}
}

func TestStreamReaderRecord(t *testing.T) {
	data := "a,b\n1,2\n3,4\n"
	it, err := StreamReader(strings.NewReader(data), WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	defer it.Close()
	it.Next()
	rec := it.Record()
	if rec["a"] != "1" || rec["b"] != "2" {
		t.Errorf("record: %v", rec)
	}
}

func TestStreamReaderFromFile(t *testing.T) {
	it, err := StreamReaderFromFile("testdata/simple.csv", WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	defer it.Close()
	count := 0
	for it.Next() {
		count++
	}
	if count != 3 {
		t.Errorf("count: %d", count)
	}
}

func TestRowIteratorRow(t *testing.T) {
	f := NewFile()
	f.AppendRow([]any{"x", "y"})
	it, _ := f.NewRowIterator()
	it.Next()
	row := it.Row()
	if row[0] != "x" || row[1] != "y" {
		t.Errorf("row: %v", row)
	}
}

func TestNewRowIteratorFromFunc(t *testing.T) {
	rows := [][]string{{"a", "1"}, {"b", "2"}, {"c", "3"}}
	idx := 0
	it := NewRowIteratorFromFunc([]string{"name", "n"}, func() ([]string, error) {
		if idx >= len(rows) {
			return nil, io.EOF
		}
		r := rows[idx]
		idx++
		return r, nil
	}, nil)
	defer it.Close()

	count := 0
	for it.Next() {
		count++
		rec := it.Record()
		if rec == nil || rec["name"] == "" {
			t.Errorf("unexpected record: %v", rec)
		}
	}
	if it.Error() != nil {
		t.Fatal(it.Error())
	}
	if count != 3 {
		t.Errorf("count: %d", count)
	}
	if it.RowIndex() != 2 {
		t.Errorf("RowIndex: %d", it.RowIndex())
	}
}

func TestNewRowIteratorFromFuncError(t *testing.T) {
	errBoom := errors.New("boom")
	it := NewRowIteratorFromFunc(nil, func() ([]string, error) {
		return nil, errBoom
	}, nil)
	defer it.Close()
	if it.Next() {
		t.Error("expected false on error")
	}
	if !errors.Is(it.Error(), errBoom) {
		t.Errorf("error: %v", it.Error())
	}
}

type countingCloser struct{ n int }

func (c *countingCloser) Close() error { c.n++; return nil }

func TestNewRowIteratorFromFuncCloser(t *testing.T) {
	c := &countingCloser{}
	it := NewRowIteratorFromFunc(nil, func() ([]string, error) {
		return nil, io.EOF
	}, c)
	for it.Next() {
	}
	if err := it.Close(); err != nil {
		t.Fatal(err)
	}
	if c.n != 1 {
		t.Errorf("closer called %d times", c.n)
	}
}
