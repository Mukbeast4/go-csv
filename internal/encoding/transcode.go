package encoding

import (
	"bytes"
	"errors"
	"io"
	"unicode/utf16"
	"unicode/utf8"
)

var ErrInvalidEncoding = errors.New("invalid encoding")

func NewDecoder(r io.Reader, enc Encoding) io.Reader {
	switch enc {
	case EncodingUTF8, EncodingAuto:
		return r
	case EncodingUTF16LE:
		return &utf16Reader{src: r, littleEndian: true}
	case EncodingUTF16BE:
		return &utf16Reader{src: r, littleEndian: false}
	case EncodingISO88591:
		return &tableReader{src: r, table: iso88591Table[:]}
	case EncodingWindows1252:
		return &tableReader{src: r, table: windows1252Table[:]}
	default:
		return r
	}
}

func NewEncoder(w io.Writer, enc Encoding, withBOM bool) io.Writer {
	switch enc {
	case EncodingUTF8, EncodingAuto:
		if withBOM {
			w.Write([]byte{0xEF, 0xBB, 0xBF})
		}
		return w
	case EncodingUTF16LE:
		if withBOM {
			w.Write([]byte{0xFF, 0xFE})
		}
		return &utf16Writer{dst: w, littleEndian: true}
	case EncodingUTF16BE:
		if withBOM {
			w.Write([]byte{0xFE, 0xFF})
		}
		return &utf16Writer{dst: w, littleEndian: false}
	case EncodingISO88591:
		return &tableWriter{dst: w, encode: encodeISO88591}
	case EncodingWindows1252:
		return &tableWriter{dst: w, encode: encodeWindows1252}
	default:
		return w
	}
}

type utf16Reader struct {
	src          io.Reader
	littleEndian bool
	buf          bytes.Buffer
	pending      []byte
	eof          bool
}

func (u *utf16Reader) Read(p []byte) (int, error) {
	for u.buf.Len() < len(p) && !u.eof {
		if err := u.fill(); err != nil {
			if err == io.EOF {
				u.eof = true
				break
			}
			return 0, err
		}
	}
	return u.buf.Read(p)
}

func (u *utf16Reader) fill() error {
	buf := make([]byte, 4096)
	n, err := u.src.Read(buf)
	if n == 0 {
		if err != nil {
			return err
		}
		return nil
	}
	data := append(u.pending, buf[:n]...)
	if len(data)%2 != 0 {
		u.pending = []byte{data[len(data)-1]}
		data = data[:len(data)-1]
	} else {
		u.pending = nil
	}
	units := make([]uint16, len(data)/2)
	for i := 0; i < len(units); i++ {
		if u.littleEndian {
			units[i] = uint16(data[i*2]) | uint16(data[i*2+1])<<8
		} else {
			units[i] = uint16(data[i*2])<<8 | uint16(data[i*2+1])
		}
	}
	runes := utf16.Decode(units)
	for _, r := range runes {
		var ubuf [4]byte
		size := utf8.EncodeRune(ubuf[:], r)
		u.buf.Write(ubuf[:size])
	}
	if err == io.EOF {
		return io.EOF
	}
	return nil
}

type utf16Writer struct {
	dst          io.Writer
	littleEndian bool
	pending      []byte
}

func (u *utf16Writer) Write(p []byte) (int, error) {
	data := append(u.pending, p...)
	consumed := 0
	out := bytes.NewBuffer(nil)
	for consumed < len(data) {
		r, size := utf8.DecodeRune(data[consumed:])
		if r == utf8.RuneError && size == 1 {
			break
		}
		if !utf8.FullRune(data[consumed:]) {
			break
		}
		consumed += size
		r1, r2 := utf16.EncodeRune(r)
		if r1 == 0xFFFD && r2 == 0xFFFD {
			writeUnit(out, uint16(r), u.littleEndian)
		} else {
			writeUnit(out, uint16(r1), u.littleEndian)
			writeUnit(out, uint16(r2), u.littleEndian)
		}
	}
	u.pending = append([]byte(nil), data[consumed:]...)
	if _, err := u.dst.Write(out.Bytes()); err != nil {
		return 0, err
	}
	return len(p), nil
}

func writeUnit(w *bytes.Buffer, u uint16, littleEndian bool) {
	if littleEndian {
		w.WriteByte(byte(u))
		w.WriteByte(byte(u >> 8))
	} else {
		w.WriteByte(byte(u >> 8))
		w.WriteByte(byte(u))
	}
}

type tableReader struct {
	src   io.Reader
	table []rune
	buf   bytes.Buffer
	eof   bool
}

func (t *tableReader) Read(p []byte) (int, error) {
	for t.buf.Len() < len(p) && !t.eof {
		buf := make([]byte, 4096)
		n, err := t.src.Read(buf)
		for i := 0; i < n; i++ {
			var ubuf [4]byte
			size := utf8.EncodeRune(ubuf[:], t.table[buf[i]])
			t.buf.Write(ubuf[:size])
		}
		if err == io.EOF {
			t.eof = true
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return t.buf.Read(p)
}

type tableWriter struct {
	dst     io.Writer
	encode  func(rune) (byte, bool)
	pending []byte
}

func (t *tableWriter) Write(p []byte) (int, error) {
	data := append(t.pending, p...)
	consumed := 0
	out := bytes.NewBuffer(nil)
	for consumed < len(data) {
		if !utf8.FullRune(data[consumed:]) {
			break
		}
		r, size := utf8.DecodeRune(data[consumed:])
		if r == utf8.RuneError && size == 1 {
			break
		}
		consumed += size
		if b, ok := t.encode(r); ok {
			out.WriteByte(b)
		} else {
			out.WriteByte('?')
		}
	}
	t.pending = append([]byte(nil), data[consumed:]...)
	if _, err := t.dst.Write(out.Bytes()); err != nil {
		return 0, err
	}
	return len(p), nil
}

func encodeISO88591(r rune) (byte, bool) {
	if r >= 0 && r <= 0xFF {
		return byte(r), true
	}
	return 0, false
}

func encodeWindows1252(r rune) (byte, bool) {
	if r >= 0 && r <= 0x7F {
		return byte(r), true
	}
	if r >= 0xA0 && r <= 0xFF {
		return byte(r), true
	}
	for i := 0x80; i <= 0x9F; i++ {
		if windows1252Table[i] == r {
			return byte(i), true
		}
	}
	return 0, false
}
