package parser

import (
	"bytes"
	"unsafe"

	"github.com/mukbeast4/go-csv/internal/dialect"
)

func ParseBytes(data []byte, d dialect.Dialect, unsafeStrings bool) ([][]string, []*ParseError, error) {
	if len(data) == 0 {
		return nil, nil, nil
	}

	delim := byte(d.Delimiter)
	quote := byte(d.Quote)
	comment := byte(d.Comment)
	asciiFast := d.Delimiter < 128 && d.Quote < 128 && (d.Comment == 0 || d.Comment < 128)
	if !asciiFast {
		return nil, nil, nil
	}

	skipped := 0
	var rows [][]string
	var errs []*ParseError
	expected := d.FieldsPerRecord
	if expected < 0 {
		expected = 0
	}
	line := 1

	pos := 0
	for pos < len(data) {
		if skipped < d.SkipRows {
			nl := bytes.IndexByte(data[pos:], '\n')
			if nl < 0 {
				return rows, errs, nil
			}
			pos += nl + 1
			skipped++
			line++
			continue
		}

		nl := bytes.IndexByte(data[pos:], '\n')
		var rawLine []byte
		var nextPos int
		if nl < 0 {
			rawLine = data[pos:]
			nextPos = len(data)
		} else {
			rawLine = data[pos : pos+nl]
			nextPos = pos + nl + 1
		}

		hasQuote := bytes.IndexByte(rawLine, quote) >= 0
		if hasQuote && nl >= 0 {
			extEnd := pos + nl
			for {
				if !quoteBalance(data[pos:extEnd], quote) {
					break
				}
				next := bytes.IndexByte(data[extEnd+1:], '\n')
				if next < 0 {
					extEnd = len(data)
					break
				}
				extEnd = extEnd + 1 + next
			}
			rawLine = data[pos:extEnd]
			if extEnd >= len(data) {
				nextPos = len(data)
			} else {
				nextPos = extEnd + 1
			}
		}

		trimmed := rawLine
		if len(trimmed) > 0 && trimmed[len(trimmed)-1] == '\r' {
			trimmed = trimmed[:len(trimmed)-1]
		}

		if comment != 0 && len(trimmed) > 0 && trimmed[0] == comment {
			pos = nextPos
			line++
			continue
		}

		if len(trimmed) == 0 {
			pos = nextPos
			line++
			continue
		}

		row, perr := parseLineBytes(trimmed, delim, quote, &d, expected, line, unsafeStrings)
		if perr != nil {
			switch d.ErrorMode {
			case dialect.ErrorModeSkip:
				errs = append(errs, perr)
				pos = nextPos
				line++
				continue
			case dialect.ErrorModeCollect:
				errs = append(errs, perr)
				if row != nil {
					rows = append(rows, row)
				}
				pos = nextPos
				line++
				continue
			default:
				return rows, errs, perr
			}
		}

		if expected == 0 {
			expected = len(row)
		} else if d.FieldsPerRecord > 0 {
			if len(row) != d.FieldsPerRecord {
				perr := &ParseError{Line: line, Err: ErrFieldCount}
				if d.ErrorMode == dialect.ErrorModeStrict {
					return rows, errs, perr
				}
				errs = append(errs, perr)
			}
		} else if d.FieldsPerRecord == -1 {
			if len(row) != expected {
				perr := &ParseError{Line: line, Err: ErrFieldCount}
				if d.ErrorMode == dialect.ErrorModeStrict {
					return rows, errs, perr
				}
				errs = append(errs, perr)
			}
		}

		rows = append(rows, row)
		pos = nextPos
		line++
	}
	return rows, errs, nil
}

func parseLineBytes(line []byte, delim, quote byte, d *dialect.Dialect, expected, lineNum int, unsafeStrings bool) ([]string, *ParseError) {
	if d.TrimLeadingSpace {
		for len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			line = line[1:]
		}
	}

	hasQuote := bytes.IndexByte(line, quote) >= 0

	if !hasQuote {
		cap := expected
		if cap == 0 {
			cap = bytes.Count(line, []byte{delim}) + 1
		}
		out := make([]string, 0, cap)
		start := 0
		for i := 0; i < len(line); i++ {
			if line[i] == delim {
				f := line[start:i]
				if d.TrimLeadingSpace {
					f = bytes.TrimLeft(f, " \t")
				}
				out = append(out, makeString(f, unsafeStrings))
				start = i + 1
			}
		}
		f := line[start:]
		if d.TrimLeadingSpace {
			f = bytes.TrimLeft(f, " \t")
		}
		out = append(out, makeString(f, unsafeStrings))
		return out, nil
	}

	cap := expected
	if cap == 0 {
		cap = 8
	}
	fields := make([]string, 0, cap)
	var fieldBuf []byte
	i := 0
	for i < len(line) {
		if d.TrimLeadingSpace {
			for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
				i++
			}
		}
		if i < len(line) && line[i] == quote {
			i++
			start := i
			closed := false
			needsCopy := false
			for i < len(line) {
				if line[i] == quote {
					if i+1 < len(line) && line[i+1] == quote {
						needsCopy = true
						break
					}
					closed = true
					break
				}
				i++
			}
			if !needsCopy && closed {
				f := line[start:i]
				fields = append(fields, makeString(f, unsafeStrings))
				i++
				if i < len(line) && line[i] == delim {
					i++
				} else if i < len(line) && !d.LazyQuotes {
					return nil, &ParseError{Line: lineNum, Err: ErrBareQuote}
				}
				continue
			}
			fieldBuf = fieldBuf[:0]
			fieldBuf = append(fieldBuf, line[start:i]...)
			for i < len(line) {
				if line[i] == quote {
					if i+1 < len(line) && line[i+1] == quote {
						fieldBuf = append(fieldBuf, quote)
						i += 2
						continue
					}
					i++
					closed = true
					break
				}
				fieldBuf = append(fieldBuf, line[i])
				i++
			}
			if !closed && !d.LazyQuotes {
				return nil, &ParseError{Line: lineNum, Err: ErrUnclosedQuote}
			}
			fields = append(fields, string(fieldBuf))
			if i < len(line) && line[i] == delim {
				i++
			} else if i < len(line) && !d.LazyQuotes {
				return nil, &ParseError{Line: lineNum, Err: ErrBareQuote}
			}
			continue
		}
		start := i
		for i < len(line) && line[i] != delim {
			if line[i] == quote && !d.LazyQuotes {
				return nil, &ParseError{Line: lineNum, Err: ErrBareQuote}
			}
			i++
		}
		fields = append(fields, makeString(line[start:i], unsafeStrings))
		if i < len(line) {
			i++
			if i == len(line) {
				fields = append(fields, "")
			}
		}
	}
	return fields, nil
}

func makeString(b []byte, useUnsafe bool) string {
	if useUnsafe {
		if len(b) == 0 {
			return ""
		}
		return unsafe.String(&b[0], len(b))
	}
	return string(b)
}
