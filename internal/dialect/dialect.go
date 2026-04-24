package dialect

type ErrorMode int

const (
	ErrorModeStrict ErrorMode = iota
	ErrorModeSkip
	ErrorModeCollect
)

type Dialect struct {
	Delimiter        rune
	Quote            rune
	Comment          rune
	LazyQuotes       bool
	TrimLeadingSpace bool
	CRLF             bool
	FieldsPerRecord  int
	SkipRows         int
	ErrorMode        ErrorMode
}

func Default() Dialect {
	return Dialect{
		Delimiter:        ',',
		Quote:            '"',
		Comment:          0,
		LazyQuotes:       false,
		TrimLeadingSpace: false,
		CRLF:             false,
		FieldsPerRecord:  0,
		SkipRows:         0,
		ErrorMode:        ErrorModeStrict,
	}
}
