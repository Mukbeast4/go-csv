package gocsv

import "testing"

func TestSniffDelimiter(t *testing.T) {
	tests := []struct {
		name string
		data string
		want rune
	}{
		{"comma", "a,b,c\n1,2,3\n", ','},
		{"semicolon", "a;b;c\n1;2;3\n", ';'},
		{"tab", "a\tb\tc\n1\t2\t3\n", '\t'},
		{"pipe", "a|b|c\n1|2|3\n", '|'},
		{"empty", "", ','},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SniffDelimiter([]byte(tt.data))
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSniffEncoding(t *testing.T) {
	if SniffEncoding([]byte{0xEF, 0xBB, 0xBF, 'a'}) != EncodingUTF8 {
		t.Error("UTF-8 BOM not detected")
	}
	if SniffEncoding([]byte{0xFF, 0xFE, 'a', 0}) != EncodingUTF16LE {
		t.Error("UTF-16LE BOM not detected")
	}
	if SniffEncoding([]byte{0xFE, 0xFF, 0, 'a'}) != EncodingUTF16BE {
		t.Error("UTF-16BE BOM not detected")
	}
}

func TestSniffHeader(t *testing.T) {
	yesHeader := [][]string{{"name", "age"}, {"Alice", "30"}}
	if !SniffHeader(yesHeader) {
		t.Error("expected header for text+numeric rows")
	}
	noHeader := [][]string{{"1", "2"}, {"3", "4"}}
	if SniffHeader(noHeader) {
		t.Error("expected no header for numeric first row")
	}
}

func TestCountUnquoted(t *testing.T) {
	got := countUnquoted(`a,b,"c,d",e`, ',')
	if got != 3 {
		t.Errorf("got %d, want 3", got)
	}
}
