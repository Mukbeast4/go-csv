package gocsv

import "testing"

func TestDefaultConfig(t *testing.T) {
	c := defaultConfig()
	if c.dialect.Delimiter != ',' {
		t.Errorf("default delimiter: got %q, want ','", c.dialect.Delimiter)
	}
	if c.dialect.Quote != '"' {
		t.Errorf("default quote: got %q", c.dialect.Quote)
	}
	if !c.autoSniff {
		t.Error("autoSniff should default to true")
	}
}

func TestOptions(t *testing.T) {
	c := applyOptions([]Option{
		WithDelimiter(';'),
		WithQuote('\''),
		WithComment('#'),
		WithHeader(true),
		WithLazyQuotes(true),
		WithTrimLeadingSpace(true),
		WithCRLF(true),
		WithEncoding(EncodingUTF16LE),
		WithErrorMode(ErrorModeSkip),
		WithFieldsPerRecord(3),
		WithSkipRows(2),
		WithBufferSize(1024),
		WithStdlibParser(),
		WithWriteBOM(true),
	})
	if c.dialect.Delimiter != ';' {
		t.Errorf("delimiter: %q", c.dialect.Delimiter)
	}
	if c.dialect.Quote != '\'' {
		t.Errorf("quote: %q", c.dialect.Quote)
	}
	if c.dialect.Comment != '#' {
		t.Errorf("comment: %q", c.dialect.Comment)
	}
	if !c.hasHeader || !c.headerSet {
		t.Error("header not set")
	}
	if !c.dialect.LazyQuotes {
		t.Error("lazy quotes not set")
	}
	if !c.dialect.TrimLeadingSpace {
		t.Error("trim not set")
	}
	if !c.dialect.CRLF {
		t.Error("CRLF not set")
	}
	if c.encoding != EncodingUTF16LE {
		t.Error("encoding not set")
	}
	if c.dialect.ErrorMode != ErrorModeSkip {
		t.Error("error mode")
	}
	if c.dialect.FieldsPerRecord != 3 {
		t.Error("FieldsPerRecord")
	}
	if c.dialect.SkipRows != 2 {
		t.Error("SkipRows")
	}
	if c.bufferSize != 1024 {
		t.Error("bufferSize")
	}
	if !c.stdlibParser {
		t.Error("stdlibParser")
	}
	if !c.writeBOM {
		t.Error("writeBOM")
	}
}

func TestWithDelimiterDisablesSniff(t *testing.T) {
	c := applyOptions([]Option{WithDelimiter(';')})
	if c.autoSniff {
		t.Error("explicit delimiter should disable autoSniff")
	}
}
