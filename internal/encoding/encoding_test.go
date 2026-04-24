package encoding

import (
	"bytes"
	"io"
	"testing"
)

func TestDetectBOM(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		enc    Encoding
		offset int
	}{
		{"utf8", []byte{0xEF, 0xBB, 0xBF, 'a'}, EncodingUTF8, 3},
		{"utf16le", []byte{0xFF, 0xFE, 'a', 0}, EncodingUTF16LE, 2},
		{"utf16be", []byte{0xFE, 0xFF, 0, 'a'}, EncodingUTF16BE, 2},
		{"no bom", []byte{'a', 'b'}, EncodingUTF8, 0},
		{"empty", []byte{}, EncodingUTF8, 0},
	}
	for _, tt := range tests {
		enc, off := DetectBOM(tt.data)
		if enc != tt.enc || off != tt.offset {
			t.Errorf("%s: got (%v,%d), want (%v,%d)", tt.name, enc, off, tt.enc, tt.offset)
		}
	}
}

func TestUTF16LEDecode(t *testing.T) {
	src := []byte{'H', 0, 'i', 0, '!', 0}
	r := NewDecoder(bytes.NewReader(src), EncodingUTF16LE)
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "Hi!" {
		t.Errorf("got %q", out)
	}
}

func TestUTF16BEDecode(t *testing.T) {
	src := []byte{0, 'H', 0, 'i'}
	r := NewDecoder(bytes.NewReader(src), EncodingUTF16BE)
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "Hi" {
		t.Errorf("got %q", out)
	}
}

func TestISO88591Decode(t *testing.T) {
	src := []byte{0x48, 0x69, 0xE9}
	r := NewDecoder(bytes.NewReader(src), EncodingISO88591)
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "Hié" {
		t.Errorf("got %q", out)
	}
}

func TestWindows1252Decode(t *testing.T) {
	src := []byte{0x80}
	r := NewDecoder(bytes.NewReader(src), EncodingWindows1252)
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "€" {
		t.Errorf("got %q", out)
	}
}

func TestUTF8EncoderWithBOM(t *testing.T) {
	var buf bytes.Buffer
	w := NewEncoder(&buf, EncodingUTF8, true)
	w.Write([]byte("hello"))
	out := buf.Bytes()
	if len(out) < 3 || out[0] != 0xEF || out[1] != 0xBB || out[2] != 0xBF {
		t.Errorf("BOM missing: %x", out)
	}
}

func TestUTF16LEEncodeDecodeRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	w := NewEncoder(&buf, EncodingUTF16LE, false)
	w.Write([]byte("Bonjour"))
	r := NewDecoder(&buf, EncodingUTF16LE)
	out, _ := io.ReadAll(r)
	if string(out) != "Bonjour" {
		t.Errorf("got %q", out)
	}
}

func TestISO88591RoundTrip(t *testing.T) {
	var buf bytes.Buffer
	w := NewEncoder(&buf, EncodingISO88591, false)
	w.Write([]byte("café"))
	r := NewDecoder(&buf, EncodingISO88591)
	out, _ := io.ReadAll(r)
	if string(out) != "café" {
		t.Errorf("got %q", out)
	}
}
