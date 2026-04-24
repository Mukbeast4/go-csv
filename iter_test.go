package gocsv

import (
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
