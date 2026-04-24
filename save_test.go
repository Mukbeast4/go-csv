package gocsv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.csv")
	f := NewFile()
	f.SetHeaders([]string{"a", "b"})
	f.AppendRow([]any{"1", "2"})
	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "a,b\n1,2\n" {
		t.Errorf("got %q", data)
	}
}

func TestWriteToBuffer(t *testing.T) {
	f := NewFile()
	f.AppendRow([]any{"x", "y"})
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "x,y\n" {
		t.Errorf("got %q", buf.String())
	}
}

func TestSaveOpenedFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.csv")
	os.WriteFile(src, []byte("a,b\n1,2\n"), 0644)
	f, err := OpenFile(src, WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	f.AppendRow([]any{"3", "4"})
	if err := f.Save(); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(src)
	if string(data) != "a,b\n1,2\n3,4\n" {
		t.Errorf("got %q", data)
	}
}

func TestWriteBOM(t *testing.T) {
	f := NewFile(WithWriteBOM(true))
	f.AppendRow([]any{"a"})
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	if len(buf.Bytes()) < 3 || buf.Bytes()[0] != 0xEF || buf.Bytes()[1] != 0xBB || buf.Bytes()[2] != 0xBF {
		t.Errorf("BOM missing: %x", buf.Bytes()[:3])
	}
}
