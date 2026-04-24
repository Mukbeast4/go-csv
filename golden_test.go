package gocsv

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoldenBasic(t *testing.T) {
	path := "testdata/golden_basic.csv"
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	f, err := OpenFile(path, WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	if !bytesEqual(original, buf.Bytes()) {
		t.Errorf("round-trip mismatch\noriginal: %q\n     got: %q", original, buf.String())
	}
}

func TestGoldenQuoted(t *testing.T) {
	path := "testdata/golden_quoted.csv"
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	f, err := OpenFile(path, WithHeader(true))
	if err != nil {
		t.Fatal(err)
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	if !bytesEqual(original, buf.Bytes()) {
		t.Errorf("round-trip mismatch\noriginal: %q\n     got: %q", original, buf.String())
	}
}

func TestGoldenAllFixtures(t *testing.T) {
	files, err := filepath.Glob("testdata/golden_*.csv")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Skip("no golden fixtures")
	}
	for _, path := range files {
		t.Run(filepath.Base(path), func(t *testing.T) {
			original, _ := os.ReadFile(path)
			f, err := OpenFile(path, WithHeader(true))
			if err != nil {
				t.Fatal(err)
			}
			buf, err := f.WriteToBuffer()
			if err != nil {
				t.Fatal(err)
			}
			if !bytesEqual(original, buf.Bytes()) {
				t.Errorf("round-trip mismatch")
			}
		})
	}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
