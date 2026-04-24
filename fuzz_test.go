package gocsv

import (
	"strings"
	"testing"
)

func FuzzParser(f *testing.F) {
	seeds := []string{
		"a,b,c\n1,2,3\n",
		`"hello","world"`,
		"a,\"b\nc\",d",
		"# comment\na,b",
		"a;b;c\n1;2;3",
		"a\tb\tc",
		"",
		"\n",
		`""""`,
		`a,"b""c",d`,
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on input %q: %v", input, r)
			}
		}()
		_, _ = OpenBytes([]byte(input),
			WithHeader(false),
			WithLazyQuotes(true),
			WithErrorMode(ErrorModeSkip),
		)
	})
}

func FuzzCoords(f *testing.F) {
	f.Add("A1")
	f.Add("Z26")
	f.Add("AA27")
	f.Add("")
	f.Add("1A")
	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on input %q: %v", input, r)
			}
		}()
		_, _, _ = CellNameToCoordinates(input)
	})
}

func FuzzRoundTrip(f *testing.F) {
	f.Add("a,b\n1,2\n")
	f.Add(`"hello","world"` + "\n")
	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on input %q: %v", input, r)
			}
		}()
		file, err := OpenBytes([]byte(input),
			WithHeader(false),
			WithLazyQuotes(true),
			WithErrorMode(ErrorModeSkip),
		)
		if err != nil {
			return
		}
		buf, err := file.WriteToBuffer()
		if err != nil {
			return
		}
		_, err = OpenBytes(buf.Bytes(),
			WithHeader(false),
			WithLazyQuotes(true),
			WithErrorMode(ErrorModeSkip),
		)
		if err != nil {
			t.Errorf("re-parse failed: input=%q output=%q err=%v", input, buf.String(), err)
		}
		_ = strings.TrimSpace
	})
}
