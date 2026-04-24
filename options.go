package gocsv

import (
	"github.com/mukbeast4/go-csv/internal/dialect"
	"github.com/mukbeast4/go-csv/internal/encoding"
)

type Encoding = encoding.Encoding

const (
	EncodingAuto        = encoding.EncodingAuto
	EncodingUTF8        = encoding.EncodingUTF8
	EncodingUTF16LE     = encoding.EncodingUTF16LE
	EncodingUTF16BE     = encoding.EncodingUTF16BE
	EncodingISO88591    = encoding.EncodingISO88591
	EncodingWindows1252 = encoding.EncodingWindows1252
)

type ErrorMode = dialect.ErrorMode

const (
	ErrorModeStrict  = dialect.ErrorModeStrict
	ErrorModeSkip    = dialect.ErrorModeSkip
	ErrorModeCollect = dialect.ErrorModeCollect
)

type config struct {
	dialect      dialect.Dialect
	encoding     Encoding
	hasHeader    bool
	headerSet    bool
	autoSniff    bool
	writeBOM     bool
	stdlibParser bool
	bufferSize   int
}

func defaultConfig() *config {
	return &config{
		dialect:    dialect.Default(),
		encoding:   EncodingAuto,
		hasHeader:  false,
		headerSet:  false,
		autoSniff:  true,
		bufferSize: 64 * 1024,
	}
}

type Option func(*config)

func WithDelimiter(r rune) Option {
	return func(c *config) {
		c.dialect.Delimiter = r
		c.autoSniff = false
	}
}

func WithQuote(r rune) Option {
	return func(c *config) { c.dialect.Quote = r }
}

func WithComment(r rune) Option {
	return func(c *config) { c.dialect.Comment = r }
}

func WithHeader(enabled bool) Option {
	return func(c *config) {
		c.hasHeader = enabled
		c.headerSet = true
	}
}

func WithEncoding(e Encoding) Option {
	return func(c *config) { c.encoding = e }
}

func WithLazyQuotes(enabled bool) Option {
	return func(c *config) { c.dialect.LazyQuotes = enabled }
}

func WithTrimLeadingSpace(enabled bool) Option {
	return func(c *config) { c.dialect.TrimLeadingSpace = enabled }
}

func WithCRLF(enabled bool) Option {
	return func(c *config) { c.dialect.CRLF = enabled }
}

func WithErrorMode(m ErrorMode) Option {
	return func(c *config) { c.dialect.ErrorMode = m }
}

func WithFieldsPerRecord(n int) Option {
	return func(c *config) { c.dialect.FieldsPerRecord = n }
}

func WithBufferSize(n int) Option {
	return func(c *config) {
		if n > 0 {
			c.bufferSize = n
		}
	}
}

func WithSkipRows(n int) Option {
	return func(c *config) {
		if n >= 0 {
			c.dialect.SkipRows = n
		}
	}
}

func WithStdlibParser() Option {
	return func(c *config) { c.stdlibParser = true }
}

func WithWriteBOM(enabled bool) Option {
	return func(c *config) { c.writeBOM = enabled }
}

func applyOptions(opts []Option) *config {
	c := defaultConfig()
	for _, opt := range opts {
		opt(c)
	}
	return c
}
