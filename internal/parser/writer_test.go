package parser

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mukbeast4/go-csv/internal/dialect"
)

func TestWriterSimple(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, dialect.Default())
	w.WriteRow([]string{"a", "b", "c"})
	w.Flush()
	if buf.String() != "a,b,c\n" {
		t.Errorf("got %q", buf.String())
	}
}

func TestWriterQuoteOnDelimiter(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, dialect.Default())
	w.WriteRow([]string{"hello, world", "bye"})
	w.Flush()
	if buf.String() != `"hello, world",bye`+"\n" {
		t.Errorf("got %q", buf.String())
	}
}

func TestWriterQuoteOnNewline(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, dialect.Default())
	w.WriteRow([]string{"multi\nline", "ok"})
	w.Flush()
	if !strings.Contains(buf.String(), `"multi`) {
		t.Errorf("got %q", buf.String())
	}
}

func TestWriterEscapeQuote(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, dialect.Default())
	w.WriteRow([]string{`she said "hi"`})
	w.Flush()
	if buf.String() != `"she said ""hi"""`+"\n" {
		t.Errorf("got %q", buf.String())
	}
}

func TestWriterCRLF(t *testing.T) {
	var buf bytes.Buffer
	d := dialect.Default()
	d.CRLF = true
	w := NewWriter(&buf, d)
	w.WriteRow([]string{"a", "b"})
	w.Flush()
	if !strings.Contains(buf.String(), "\r\n") {
		t.Errorf("expected CRLF: %q", buf.String())
	}
}

func TestWriterWriteAll(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, dialect.Default())
	w.WriteAll([][]string{{"a", "b"}, {"1", "2"}})
	if buf.String() != "a,b\n1,2\n" {
		t.Errorf("got %q", buf.String())
	}
}

func TestWriterEmpty(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf, dialect.Default())
	w.WriteRow([]string{})
	w.Flush()
	if buf.String() != "\n" {
		t.Errorf("got %q", buf.String())
	}
}
