package gocsv

import (
	"strings"
	"testing"
)

func TestNewFile(t *testing.T) {
	f := NewFile()
	if f.RowCount() != 0 {
		t.Error("new file should have 0 rows")
	}
}

func TestOpenBytesSimple(t *testing.T) {
	data := []byte("a,b,c\n1,2,3\n4,5,6\n")
	f, err := OpenBytes(data, WithHeader(false))
	if err != nil {
		t.Fatal(err)
	}
	if f.RowCount() != 3 {
		t.Errorf("rows: got %d, want 3", f.RowCount())
	}
	if f.ColCount() != 3 {
		t.Errorf("cols: got %d, want 3", f.ColCount())
	}
}

func TestOpenBytesWithHeader(t *testing.T) {
	data := []byte("name,age\nAlice,30\nBob,25\n")
	f, err := OpenBytes(data, WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	if !f.HasHeader() {
		t.Error("should have header")
	}
	if f.RowCount() != 2 {
		t.Errorf("rows: got %d, want 2", f.RowCount())
	}
	headers := f.Headers()
	if len(headers) != 2 || headers[0] != "name" || headers[1] != "age" {
		t.Errorf("headers: %v", headers)
	}
}

func TestOpenFileSimple(t *testing.T) {
	f, err := OpenFile("testdata/simple.csv", WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	if f.RowCount() != 3 {
		t.Errorf("rows: %d", f.RowCount())
	}
	if f.Path() != "testdata/simple.csv" {
		t.Errorf("path: %s", f.Path())
	}
}

func TestOpenReaderSimple(t *testing.T) {
	r := strings.NewReader("x,y\n1,2\n")
	f, err := OpenReader(r, WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	if f.RowCount() != 1 {
		t.Errorf("rows: %d", f.RowCount())
	}
}

func TestAutoSniffDelimiter(t *testing.T) {
	data := []byte("a;b;c\n1;2;3\n4;5;6\n")
	f, err := OpenBytes(data, WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	if f.ColCount() != 3 {
		t.Errorf("auto-sniff should detect ';': got %d cols", f.ColCount())
	}
}

func TestOpenFileQuoted(t *testing.T) {
	f, err := OpenFile("testdata/quoted.csv", WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	if f.RowCount() != 3 {
		t.Fatalf("rows: %d", f.RowCount())
	}
	rows, _ := f.GetRows()
	if rows[0][1] != "123 Main St, Apt 4" {
		t.Errorf("embedded comma failed: %q", rows[0][1])
	}
	if rows[0][2] != `She said "hello"` {
		t.Errorf("embedded quote failed: %q", rows[0][2])
	}
	if !strings.Contains(rows[1][2], "\n") {
		t.Errorf("embedded newline failed: %q", rows[1][2])
	}
}

func TestDimension(t *testing.T) {
	f := NewFile()
	f.AppendRow([]any{"a", "b", "c"})
	f.AppendRow([]any{"1", "2", "3"})
	if f.Dimension() != "A1:C2" {
		t.Errorf("dimension: %q", f.Dimension())
	}
}

func TestCloseIdempotent(t *testing.T) {
	f := NewFile()
	if err := f.Close(); err != nil {
		t.Error(err)
	}
	if err := f.Close(); err != nil {
		t.Error(err)
	}
	if _, err := f.GetRows(); err != ErrFileClosed {
		t.Errorf("expected ErrFileClosed, got %v", err)
	}
}

func TestStdlibParserOption(t *testing.T) {
	data := []byte("a,b\n1,2\n")
	f, err := OpenBytes(data, WithStdlibParser(), WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	if f.RowCount() != 1 {
		t.Errorf("rows: %d", f.RowCount())
	}
}
