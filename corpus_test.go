package gocsv

import (
	"os"
	"strings"
	"testing"
)

func TestCorpusSimple(t *testing.T) {
	f, err := OpenFile("testdata/simple.csv", WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	if f.RowCount() != 3 {
		t.Errorf("rows: %d", f.RowCount())
	}
}

func TestCorpusQuoted(t *testing.T) {
	f, err := OpenFile("testdata/quoted.csv", WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	rows, _ := f.GetRows()
	for _, r := range rows {
		if len(r) != 3 {
			t.Errorf("row should have 3 fields: %v", r)
		}
	}
}

func TestCorpusSemicolon(t *testing.T) {
	f, err := OpenFile("testdata/semicolon.csv", WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	if f.ColCount() != 3 {
		t.Errorf("cols: %d", f.ColCount())
	}
}

func TestCorpusTSV(t *testing.T) {
	f, err := OpenFile("testdata/tsv.csv", WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	if f.ColCount() != 3 {
		t.Errorf("cols: %d", f.ColCount())
	}
}

func TestCorpusComments(t *testing.T) {
	f, err := OpenFile("testdata/comments.csv", WithHeader(true), WithComment('#'))
	if err != nil {
		t.Fatal(err)
	}
	if f.RowCount() != 2 {
		t.Errorf("rows: %d", f.RowCount())
	}
}

func TestCorpusCRLF(t *testing.T) {
	f, err := OpenFile("testdata/crlf.csv", WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	if f.RowCount() != 2 {
		t.Errorf("rows: %d", f.RowCount())
	}
	rows, _ := f.GetRows()
	for _, r := range rows {
		for _, field := range r {
			if strings.Contains(field, "\r") {
				t.Errorf("CRLF not stripped: %q", field)
			}
		}
	}
}

func TestCorpusUTF8BOM(t *testing.T) {
	f, err := OpenFile("testdata/utf8_bom.csv", WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	headers := f.Headers()
	if len(headers) == 0 {
		t.Fatal("no headers")
	}
	if strings.Contains(headers[0], "\uFEFF") {
		t.Errorf("BOM not stripped: %q", headers[0])
	}
	if headers[0] != "name" {
		t.Errorf("first header: %q, want 'name'", headers[0])
	}
}

func TestCorpusEmptyFile(t *testing.T) {
	data, _ := os.ReadFile("testdata/simple.csv")
	_ = data
	f, err := OpenBytes([]byte(""), WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	if f.RowCount() != 0 {
		t.Error("empty file should have 0 rows")
	}
}

func TestCorpusSingleRowNoNewline(t *testing.T) {
	f, err := OpenBytes([]byte("a,b,c"), WithHeader(false))
	if err != nil {
		t.Fatal(err)
	}
	if f.RowCount() != 1 {
		t.Errorf("rows: %d", f.RowCount())
	}
}

func TestCorpusRagged(t *testing.T) {
	f, err := OpenFile("testdata/ragged.csv", WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	rows, _ := f.GetRows()
	if len(rows) != 3 {
		t.Errorf("rows: %d", len(rows))
	}
}
