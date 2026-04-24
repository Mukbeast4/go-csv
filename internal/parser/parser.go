package parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/mukbeast4/go-csv/internal/dialect"
)

var (
	ErrBareQuote     = errors.New("bare quote in unquoted field")
	ErrUnclosedQuote = errors.New("unclosed quoted field")
	ErrFieldCount    = errors.New("row field count mismatch")
)

type ParseError struct {
	Line   int
	Column int
	Offset int64
	Err    error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d column %d (offset %d): %v", e.Line, e.Column, e.Offset, e.Err)
}

func (e *ParseError) Unwrap() error { return e.Err }

type Parser struct {
	r         *bufio.Reader
	d         dialect.Dialect
	delim     byte
	quote     byte
	comment   byte
	line      int
	offset    int64
	fields    []string
	fieldBuf  []byte
	done      bool
	skipped   int
	expected  int
	asciiFast bool
}

func New(r io.Reader, d dialect.Dialect) *Parser {
	p := &Parser{
		r:         bufio.NewReaderSize(r, 128*1024),
		d:         d,
		line:      1,
		expected:  d.FieldsPerRecord,
		asciiFast: d.Delimiter < 128 && d.Quote < 128 && (d.Comment == 0 || d.Comment < 128),
	}
	if p.asciiFast {
		p.delim = byte(d.Delimiter)
		p.quote = byte(d.Quote)
		if d.Comment != 0 {
			p.comment = byte(d.Comment)
		}
	}
	return p
}

func (p *Parser) Line() int     { return p.line }
func (p *Parser) Offset() int64 { return p.offset }
func (p *Parser) Err() error    { return nil }
func (p *Parser) Done() bool    { return p.done }

func (p *Parser) Next() ([]string, error) {
	if p.done {
		return nil, io.EOF
	}
	for p.skipped < p.d.SkipRows {
		if err := p.skipLine(); err != nil {
			p.done = true
			return nil, err
		}
		p.skipped++
	}
	for {
		row, err := p.readRecord()
		if err == io.EOF {
			p.done = true
			return nil, io.EOF
		}
		if err != nil {
			return row, err
		}
		if p.comment != 0 && len(row) > 0 && len(row[0]) > 0 && row[0][0] == p.comment {
			continue
		}
		if p.expected == 0 {
			p.expected = len(row)
		} else if p.d.FieldsPerRecord > 0 {
			if len(row) != p.d.FieldsPerRecord {
				return row, p.parseErr(ErrFieldCount)
			}
		} else if p.d.FieldsPerRecord == -1 {
			if len(row) != p.expected {
				return row, p.parseErr(ErrFieldCount)
			}
		}
		return row, nil
	}
}

func (p *Parser) skipLine() error {
	_, err := p.r.ReadSlice('\n')
	if err != nil && err != bufio.ErrBufferFull {
		if err == io.EOF {
			return err
		}
		return err
	}
	p.line++
	return nil
}

func (p *Parser) parseErr(err error) *ParseError {
	return &ParseError{
		Line:   p.line,
		Offset: p.offset,
		Err:    err,
	}
}

func (p *Parser) readRecord() ([]string, error) {
	p.fields = p.fields[:0]

	line, hasQuote, err := p.readLine()
	if err != nil {
		if len(line) == 0 {
			return nil, io.EOF
		}
	}

	if !hasQuote {
		return p.parseLineSimple(line, err)
	}

	return p.parseLineComplex(line, err)
}

func (p *Parser) readLine() ([]byte, bool, error) {
	line, err := p.r.ReadSlice('\n')
	p.offset += int64(len(line))

	hasQuote := bytes.IndexByte(line, p.quote) >= 0
	if err == nil || err == io.EOF {
		if !hasQuote {
			return line, false, err
		}
		if !quoteBalance(line, p.quote) {
			return line, true, err
		}
	}

	full := append([]byte(nil), line...)
	inQuote := hasQuote && quoteBalance(full, p.quote)
	for inQuote || err == bufio.ErrBufferFull {
		if err != nil && err != bufio.ErrBufferFull {
			break
		}
		more, merr := p.r.ReadSlice('\n')
		p.offset += int64(len(more))
		full = append(full, more...)
		err = merr
		if bytes.IndexByte(more, p.quote) >= 0 {
			hasQuote = true
		}
		if hasQuote {
			inQuote = quoteBalance(full, p.quote)
		} else {
			inQuote = false
		}
		if merr == io.EOF {
			break
		}
	}
	return full, hasQuote, err
}

func quoteBalance(line []byte, quote byte) bool {
	count := 0
	for i := 0; i < len(line); i++ {
		if line[i] == quote {
			count++
		}
	}
	return count%2 == 1
}

func (p *Parser) parseLineSimple(line []byte, readErr error) ([]string, error) {
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
	}

	if p.d.TrimLeadingSpace {
		line = trimLeadingHeadSpace(line)
	}

	if len(line) == 0 && readErr == io.EOF {
		return nil, io.EOF
	}

	expected := p.expected
	if expected == 0 {
		expected = bytes.Count(line, []byte{p.delim}) + 1
	}
	out := make([]string, 0, expected)

	start := 0
	for i := 0; i < len(line); i++ {
		if line[i] == p.delim {
			f := line[start:i]
			if p.d.TrimLeadingSpace {
				f = bytes.TrimLeft(f, " \t")
			}
			out = append(out, string(f))
			start = i + 1
		}
	}
	f := line[start:]
	if p.d.TrimLeadingSpace {
		f = bytes.TrimLeft(f, " \t")
	}
	out = append(out, string(f))

	p.line++
	return out, nil
}

func trimLeadingHeadSpace(b []byte) []byte {
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\t') {
		b = b[1:]
	}
	return b
}

func (p *Parser) parseLineComplex(line []byte, readErr error) ([]string, error) {
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
	}

	if len(line) == 0 && readErr == io.EOF {
		return nil, io.EOF
	}

	expected := p.expected
	if expected == 0 {
		expected = 8
	}
	fields := make([]string, 0, expected)
	p.fieldBuf = p.fieldBuf[:0]
	i := 0
	for i < len(line) {
		if p.d.TrimLeadingSpace {
			for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
				i++
			}
		}
		if i < len(line) && line[i] == p.quote {
			i++
			closed := false
			for i < len(line) {
				if line[i] == p.quote {
					if i+1 < len(line) && line[i+1] == p.quote {
						p.fieldBuf = append(p.fieldBuf, p.quote)
						i += 2
						continue
					}
					i++
					closed = true
					break
				}
				p.fieldBuf = append(p.fieldBuf, line[i])
				i++
			}
			if !closed && !p.d.LazyQuotes {
				return nil, p.parseErr(ErrUnclosedQuote)
			}
			if i < len(line) && line[i] == p.delim {
				i++
			} else if i < len(line) {
				if !p.d.LazyQuotes {
					return nil, p.parseErr(ErrBareQuote)
				}
			}
			fields = append(fields, string(p.fieldBuf))
			p.fieldBuf = p.fieldBuf[:0]
			continue
		}
		start := i
		for i < len(line) && line[i] != p.delim {
			if line[i] == p.quote && !p.d.LazyQuotes {
				return nil, p.parseErr(ErrBareQuote)
			}
			i++
		}
		fields = append(fields, string(line[start:i]))
		if i < len(line) {
			i++
			if i == len(line) {
				fields = append(fields, "")
			}
		}
	}
	p.line++
	return fields, nil
}

func ReadAll(r io.Reader, d dialect.Dialect) ([][]string, error) {
	p := New(r, d)
	var rows [][]string
	for {
		row, err := p.Next()
		if err == io.EOF {
			return rows, nil
		}
		if err != nil {
			if d.ErrorMode == dialect.ErrorModeSkip {
				continue
			}
			if d.ErrorMode == dialect.ErrorModeCollect {
				rows = append(rows, row)
				continue
			}
			return rows, err
		}
		rows = append(rows, row)
	}
}
