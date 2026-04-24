package ods

import (
	"os"
	"path/filepath"
	"testing"

	gocsv "github.com/mukbeast4/go-csv"
)

func sampleCSV() *gocsv.File {
	f := gocsv.NewFile()
	f.SetHeaders([]string{"id", "name", "score"})
	f.AppendRow([]any{1, "Alice", 95})
	f.AppendRow([]any{2, "Bob", 87})
	f.AppendRow([]any{3, "Charlie", 72})
	return f
}

func TestToODS(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.ods")

	if err := ToODS(sampleCSV(), path); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Error("empty file")
	}
}

func TestFromODS(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rt.ods")

	if err := ToODS(sampleCSV(), path, WithSheetName("Data")); err != nil {
		t.Fatal(err)
	}

	out, err := FromODS(path, "Data")
	if err != nil {
		t.Fatal(err)
	}
	if out.RowCount() != 3 {
		t.Errorf("rows: %d", out.RowCount())
	}
	headers := out.Headers()
	if len(headers) != 3 || headers[0] != "id" || headers[1] != "name" {
		t.Errorf("headers: %v", headers)
	}
}

func TestRoundTripData(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rt2.ods")

	csv := sampleCSV()
	if err := ToODS(csv, path); err != nil {
		t.Fatal(err)
	}
	out, err := FromODS(path, "Sheet1")
	if err != nil {
		t.Fatal(err)
	}
	rows, _ := out.GetRows()
	if len(rows) != 3 {
		t.Fatalf("rows: %d", len(rows))
	}
	if rows[0][1] != "Alice" {
		t.Errorf("alice: %v", rows[0])
	}
}

func TestWithoutHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "noh.ods")

	f := gocsv.NewFile()
	f.AppendRow([]any{"a", "b"})
	f.AppendRow([]any{"c", "d"})

	if err := ToODS(f, path, WithHeader(false)); err != nil {
		t.Fatal(err)
	}
	out, err := FromODS(path, "Sheet1", WithHeader(false))
	if err != nil {
		t.Fatal(err)
	}
	if out.RowCount() != 2 {
		t.Errorf("rows: %d", out.RowCount())
	}
}
