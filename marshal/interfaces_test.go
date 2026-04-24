package marshal

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"
)

type color struct{ R, G, B int }

func (c color) MarshalCSV() (string, error) {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B), nil
}

func (c *color) UnmarshalCSV(s string) error {
	if !strings.HasPrefix(s, "#") || len(s) != 7 {
		return fmt.Errorf("invalid color: %q", s)
	}
	var r, g, b int
	if _, err := fmt.Sscanf(s[1:], "%02x%02x%02x", &r, &g, &b); err != nil {
		return err
	}
	c.R, c.G, c.B = r, g, b
	return nil
}

type paletteRow struct {
	Name    string `csv:"name"`
	Primary color  `csv:"primary"`
}

func TestCSVMarshalerRoundTrip(t *testing.T) {
	items := []paletteRow{
		{Name: "red", Primary: color{R: 255, G: 0, B: 0}},
		{Name: "teal", Primary: color{R: 0, G: 128, B: 128}},
	}
	data, err := Marshal(items)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "#ff0000") {
		t.Errorf("MarshalCSV not used: %q", data)
	}
	var got []paletteRow
	if err := Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Primary.R != 255 || got[1].Primary.G != 128 {
		t.Errorf("round-trip mismatch: %+v", got)
	}
}

type errCell struct{}

func (errCell) MarshalCSV() (string, error) { return "", errors.New("marshal-fail") }

func TestCSVMarshalerErrorPropagates(t *testing.T) {
	type R struct {
		X errCell `csv:"x"`
	}
	_, err := Marshal([]R{{}})
	if err == nil || !strings.Contains(err.Error(), "marshal-fail") {
		t.Errorf("want marshal-fail, got %v", err)
	}
}

type bigIntRow struct {
	ID  int      `csv:"id"`
	Amt *big.Int `csv:"amt"`
}

func TestTextMarshalerFallback(t *testing.T) {
	rows := []bigIntRow{{ID: 1, Amt: big.NewInt(12345)}}
	data, err := Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "12345") {
		t.Errorf("big.Int not encoded via TextMarshaler: %q", data)
	}
	var back []bigIntRow
	if err := Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if len(back) != 1 || back[0].Amt == nil || back[0].Amt.Int64() != 12345 {
		t.Errorf("got: %+v", back)
	}
}

type bothCell struct{ v string }

func (b bothCell) MarshalCSV() (string, error)  { return "csv:" + b.v, nil }
func (b bothCell) MarshalText() ([]byte, error) { return []byte("text:" + b.v), nil }
func (b *bothCell) UnmarshalCSV(s string) error { b.v = strings.TrimPrefix(s, "csv:"); return nil }
func (b *bothCell) UnmarshalText(data []byte) error {
	b.v = strings.TrimPrefix(string(data), "text:")
	return nil
}

func TestCSVMarshalerPriorityOverText(t *testing.T) {
	type R struct {
		X bothCell `csv:"x"`
	}
	rows := []R{{X: bothCell{v: "hello"}}}
	data, err := Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "csv:hello") {
		t.Errorf("expected CSV path, got: %q", data)
	}
}

func TestTimeNotHijackedByInterface(t *testing.T) {
	data, err := Marshal(sampleUsers())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "2025-01-15") {
		t.Errorf("time format=2006-01-02 lost: %q", data)
	}
}
