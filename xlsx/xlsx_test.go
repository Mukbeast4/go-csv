package xlsx

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	gocsv "github.com/mukbeast4/go-csv"
)

func sampleFile() *gocsv.File {
	f := gocsv.NewFile()
	f.SetHeaders([]string{"id", "name", "active"})
	f.AppendStrRow([]string{"1", "Alice", "true"})
	f.AppendStrRow([]string{"2", "Bob", "false"})
	return f
}

func TestToXLSXBasic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.xlsx")
	if err := ToXLSX(sampleFile(), path); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Error("empty file")
	}

	back, err := FromXLSX(path, "")
	if err != nil {
		t.Fatal(err)
	}
	if back.RowCount() != 2 {
		t.Errorf("rows: %d", back.RowCount())
	}
	if got, _ := back.GetByHeader(0, "name"); got != "Alice" {
		t.Errorf("name: %q", got)
	}
}

func TestToXLSXCustomSheetName(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.xlsx")
	if err := ToXLSX(sampleFile(), path, WithSheetName("Users")); err != nil {
		t.Fatal(err)
	}
	names, err := SheetNames(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "Users" {
		t.Errorf("sheet names: %v", names)
	}
}

func TestAppendSheet(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "multi.xlsx")
	if err := ToXLSX(sampleFile(), path, WithSheetName("Users")); err != nil {
		t.Fatal(err)
	}
	f2 := gocsv.NewFile()
	f2.SetHeaders([]string{"country"})
	f2.AppendStrRow([]string{"FR"})
	if err := AppendSheet(path, "Countries", f2); err != nil {
		t.Fatal(err)
	}
	names, _ := SheetNames(path)
	seen := map[string]bool{}
	for _, n := range names {
		seen[n] = true
	}
	if !seen["Users"] || !seen["Countries"] {
		t.Errorf("missing sheets: %v", names)
	}
}

func TestWriteReadXLSXBuffer(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteXLSX(sampleFile(), &buf); err != nil {
		t.Fatal(err)
	}
	back, err := ReadXLSX(&buf, "")
	if err != nil {
		t.Fatal(err)
	}
	if back.RowCount() != 2 {
		t.Errorf("rows: %d", back.RowCount())
	}
}

func TestStreamWriteRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "stream.xlsx")
	sw, closer, err := NewStreamWriter(path, "Data")
	if err != nil {
		t.Fatal(err)
	}
	if err := sw.WriteHeader([]string{"id", "name"}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 500; i++ {
		if err := sw.WriteStrRow([]string{"1", "x"}); err != nil {
			t.Fatal(err)
		}
	}
	if err := closer.Close(); err != nil {
		t.Fatal(err)
	}

	it, rcloser, err := NewStreamReader(path, "Data")
	if err != nil {
		t.Fatal(err)
	}
	defer rcloser.Close()
	count := 0
	for it.Next() {
		count++
	}
	if it.Error() != nil {
		t.Fatal(it.Error())
	}
	if count != 500 {
		t.Errorf("count: %d", count)
	}
}

func TestHeaderOptionFalse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "noheader.xlsx")
	if err := ToXLSX(sampleFile(), path, WithHeader(false)); err != nil {
		t.Fatal(err)
	}
	back, err := FromXLSX(path, "", WithHeader(false))
	if err != nil {
		t.Fatal(err)
	}
	// When no header written, the 3 header strings appear as data rows in addition
	// to the 2 data rows. That's 3 + 2 - but headers are never serialized so it's
	// just 2 rows of original data. Actually when header=false on write, headers
	// are NOT written at all, so the output has exactly 2 data rows.
	if back.RowCount() != 2 {
		t.Errorf("rows: %d (headers=%v)", back.RowCount(), back.Headers())
	}
}
