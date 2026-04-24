package parser

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/mukbeast4/go-csv/internal/dialect"
)

func rows(t *testing.T, input string, d dialect.Dialect) [][]string {
	t.Helper()
	rows, err := ReadAll(strings.NewReader(input), d)
	if err != nil {
		t.Fatalf("ReadAll error: %v", err)
	}
	return rows
}

func TestParserSimple(t *testing.T) {
	got := rows(t, "a,b,c\n1,2,3\n", dialect.Default())
	if len(got) != 2 || got[0][0] != "a" || got[1][2] != "3" {
		t.Errorf("got %v", got)
	}
}

func TestParserEmptyLines(t *testing.T) {
	got := rows(t, "a,b\n\n1,2\n", dialect.Default())
	if len(got) < 2 {
		t.Errorf("got %v", got)
	}
}

func TestParserQuoted(t *testing.T) {
	got := rows(t, `a,"b,c",d`+"\n", dialect.Default())
	if got[0][1] != "b,c" {
		t.Errorf("got %q", got[0][1])
	}
}

func TestParserEscapedQuote(t *testing.T) {
	got := rows(t, `a,"b""c",d`+"\n", dialect.Default())
	if got[0][1] != `b"c` {
		t.Errorf("got %q", got[0][1])
	}
}

func TestParserEmbeddedNewline(t *testing.T) {
	got := rows(t, "a,\"b\nc\",d\n", dialect.Default())
	if got[0][1] != "b\nc" {
		t.Errorf("got %q", got[0][1])
	}
}

func TestParserCRLF(t *testing.T) {
	got := rows(t, "a,b\r\n1,2\r\n", dialect.Default())
	if len(got) != 2 {
		t.Errorf("got %v", got)
	}
}

func TestParserMixedLineEndings(t *testing.T) {
	got := rows(t, "a,b\n1,2\r\n3,4\n", dialect.Default())
	if len(got) != 3 {
		t.Errorf("got %v", got)
	}
}

func TestParserTabDelimiter(t *testing.T) {
	d := dialect.Default()
	d.Delimiter = '\t'
	got := rows(t, "a\tb\n1\t2\n", d)
	if got[0][0] != "a" || got[0][1] != "b" {
		t.Errorf("got %v", got)
	}
}

func TestParserSemicolonDelimiter(t *testing.T) {
	d := dialect.Default()
	d.Delimiter = ';'
	got := rows(t, "a;b\n1;2\n", d)
	if got[1][1] != "2" {
		t.Errorf("got %v", got)
	}
}

func TestParserTrimLeading(t *testing.T) {
	d := dialect.Default()
	d.TrimLeadingSpace = true
	got := rows(t, "  a, b , c\n", d)
	if got[0][0] != "a" || got[0][1] != "b " {
		t.Errorf("got %v", got)
	}
}

func TestParserComment(t *testing.T) {
	d := dialect.Default()
	d.Comment = '#'
	got := rows(t, "# comment\na,b\n1,2\n", d)
	if len(got) != 2 {
		t.Errorf("got %v", got)
	}
}

func TestParserBareQuote(t *testing.T) {
	_, err := ReadAll(strings.NewReader(`a,"b"c`+"\n"), dialect.Default())
	if err == nil {
		t.Error("expected bare quote error")
	}
	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Errorf("expected *ParseError, got %T", err)
	}
}

func TestParserLazyQuotes(t *testing.T) {
	d := dialect.Default()
	d.LazyQuotes = true
	got, err := ReadAll(strings.NewReader(`a,"b"c,d`+"\n"), d)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Errorf("got %v", got)
	}
}

func TestParserUnclosedQuote(t *testing.T) {
	_, err := ReadAll(strings.NewReader(`"unclosed`), dialect.Default())
	if err == nil {
		t.Error("expected unclosed quote error")
	}
}

func TestParserNoTrailingNewline(t *testing.T) {
	got := rows(t, "a,b\n1,2", dialect.Default())
	if len(got) != 2 || got[1][1] != "2" {
		t.Errorf("got %v", got)
	}
}

func TestParserEmpty(t *testing.T) {
	p := New(strings.NewReader(""), dialect.Default())
	_, err := p.Next()
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
}

func TestParserSkipRows(t *testing.T) {
	d := dialect.Default()
	d.SkipRows = 2
	got := rows(t, "skip1\nskip2\na,b\n1,2\n", d)
	if len(got) != 2 || got[0][0] != "a" {
		t.Errorf("got %v", got)
	}
}

func TestParserFieldsPerRecord(t *testing.T) {
	d := dialect.Default()
	d.FieldsPerRecord = 2
	_, err := ReadAll(strings.NewReader("a,b\n1,2,3\n"), d)
	if err == nil {
		t.Error("expected field count error")
	}
}

func TestParserErrorModeSkip(t *testing.T) {
	d := dialect.Default()
	d.ErrorMode = dialect.ErrorModeSkip
	got, err := ReadAll(strings.NewReader(`a,"unclosed`), d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = got
}
