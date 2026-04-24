package gocsv

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStreamWriter(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)
	sw.WriteHeader([]string{"a", "b"})
	sw.WriteRow([]any{"1", "2"})
	sw.WriteRow([]any{"3", "4"})
	if err := sw.Close(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	expected := "a,b\n1,2\n3,4\n"
	if out != expected {
		t.Errorf("got %q, want %q", out, expected)
	}
}

func TestStreamWriterRecord(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)
	sw.WriteHeader([]string{"name", "age"})
	sw.WriteRecord(map[string]any{"name": "Alice", "age": 30})
	sw.WriteRecord(map[string]any{"name": "Bob", "age": 25})
	sw.Close()
	if !strings.Contains(buf.String(), "Alice,30") || !strings.Contains(buf.String(), "Bob,25") {
		t.Errorf("got %q", buf.String())
	}
}

func TestStreamWriterCRLF(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf, WithCRLF(true))
	sw.WriteHeader([]string{"a"})
	sw.WriteRow([]any{"1"})
	sw.Close()
	if !strings.Contains(buf.String(), "\r\n") {
		t.Errorf("expected CRLF: %q", buf.String())
	}
}

func TestStreamWriterQuoting(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)
	sw.WriteRow([]any{"hello, world", `she said "hi"`, "line1\nline2"})
	sw.Close()
	s := buf.String()
	if !strings.Contains(s, `"hello, world"`) {
		t.Errorf("embedded comma not quoted: %q", s)
	}
	if !strings.Contains(s, `"she said ""hi"""`) {
		t.Errorf("embedded quote not escaped: %q", s)
	}
}

func TestStreamWriterToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.csv")
	sw, err := NewStreamWriterToFile(path)
	if err != nil {
		t.Fatal(err)
	}
	sw.WriteRow([]any{"a", "b"})
	sw.Close()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "a,b\n" {
		t.Errorf("got %q", data)
	}
}

func TestStreamWriterClosedError(t *testing.T) {
	var buf bytes.Buffer
	sw := NewStreamWriter(&buf)
	sw.Close()
	if err := sw.WriteRow([]any{"a"}); err != ErrStreamClosed {
		t.Errorf("expected ErrStreamClosed, got %v", err)
	}
}
