package compress

import (
	"os"
	"path/filepath"
	"testing"

	gocsv "github.com/mukbeast4/go-csv"
)

func TestDetectFormat(t *testing.T) {
	tests := map[string]Format{
		"file.csv":    FormatNone,
		"file.csv.gz": FormatGzip,
		"file.csv.bz2": FormatBzip2,
		"file.CSV.GZ": FormatGzip,
	}
	for path, want := range tests {
		if got := DetectFormat(path); got != want {
			t.Errorf("%s: got %v want %v", path, got, want)
		}
	}
}

func TestGzipRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.csv.gz")

	f := gocsv.NewFile()
	f.SetHeaders([]string{"id", "name"})
	f.AppendRow([]any{1, "Alice"})
	f.AppendRow([]any{2, "Bob"})
	if err := SaveAs(f, path); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Error("empty file")
	}

	opened, err := Open(path, gocsv.WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	if opened.RowCount() != 2 {
		t.Errorf("rows: %d", opened.RowCount())
	}
}

func TestStreamWriterGzip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stream.csv.gz")

	sw, closer, err := NewStreamWriter(path)
	if err != nil {
		t.Fatal(err)
	}
	sw.WriteHeader([]string{"id", "val"})
	for i := 0; i < 100; i++ {
		sw.WriteRow([]any{i, "x"})
	}
	if err := closer.Close(); err != nil {
		t.Fatal(err)
	}

	opened, err := Open(path, gocsv.WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	if opened.RowCount() != 100 {
		t.Errorf("rows: %d", opened.RowCount())
	}
}

func TestStreamReaderGzip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "read.csv.gz")

	f := gocsv.NewFile()
	f.SetHeaders([]string{"id"})
	for i := 0; i < 50; i++ {
		f.AppendRow([]any{i})
	}
	SaveAs(f, path)

	it, closer, err := NewStreamReader(path, gocsv.WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	defer closer.Close()

	count := 0
	for it.Next() {
		count++
	}
	if count != 50 {
		t.Errorf("count: %d", count)
	}
}

func TestOpenPlainFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plain.csv")
	os.WriteFile(path, []byte("a,b\n1,2\n"), 0644)

	f, err := Open(path, gocsv.WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	if f.RowCount() != 1 {
		t.Errorf("rows: %d", f.RowCount())
	}
}
