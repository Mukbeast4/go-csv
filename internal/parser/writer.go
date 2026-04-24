package parser

import (
	"bufio"
	"io"
	"strings"

	"github.com/mukbeast4/go-csv/internal/dialect"
)

type Writer struct {
	w   *bufio.Writer
	d   dialect.Dialect
	err error
}

func NewWriter(w io.Writer, d dialect.Dialect) *Writer {
	return &Writer{
		w: bufio.NewWriterSize(w, 64*1024),
		d: d,
	}
}

func (w *Writer) WriteRow(row []string) error {
	if w.err != nil {
		return w.err
	}
	for i, field := range row {
		if i > 0 {
			if _, err := w.w.WriteRune(w.d.Delimiter); err != nil {
				w.err = err
				return err
			}
		}
		if err := w.writeField(field); err != nil {
			w.err = err
			return err
		}
	}
	if w.d.CRLF {
		if _, err := w.w.WriteString("\r\n"); err != nil {
			w.err = err
			return err
		}
	} else {
		if err := w.w.WriteByte('\n'); err != nil {
			w.err = err
			return err
		}
	}
	return nil
}

func (w *Writer) WriteAll(rows [][]string) error {
	for _, row := range rows {
		if err := w.WriteRow(row); err != nil {
			return err
		}
	}
	return w.Flush()
}

func (w *Writer) Flush() error {
	if w.err != nil {
		return w.err
	}
	if err := w.w.Flush(); err != nil {
		w.err = err
		return err
	}
	return nil
}

func (w *Writer) Error() error { return w.err }

func (w *Writer) writeField(field string) error {
	if !w.needsQuote(field) {
		_, err := w.w.WriteString(field)
		return err
	}
	if _, err := w.w.WriteRune(w.d.Quote); err != nil {
		return err
	}
	quoteStr := string(w.d.Quote)
	escaped := strings.ReplaceAll(field, quoteStr, quoteStr+quoteStr)
	if _, err := w.w.WriteString(escaped); err != nil {
		return err
	}
	if _, err := w.w.WriteRune(w.d.Quote); err != nil {
		return err
	}
	return nil
}

func (w *Writer) needsQuote(field string) bool {
	if field == "" {
		return false
	}
	for _, r := range field {
		if r == w.d.Delimiter || r == w.d.Quote || r == '\n' || r == '\r' {
			return true
		}
	}
	return false
}
