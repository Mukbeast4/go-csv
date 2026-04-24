package jsonx

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	gocsv "github.com/mukbeast4/go-csv"
)

func sampleFile() *gocsv.File {
	f := gocsv.NewFile()
	f.SetHeaders([]string{"id", "name", "active"})
	f.AppendStrRow([]string{"1", "Alice", "true"})
	f.AppendStrRow([]string{"2", "Bob", "false"})
	return f
}

func TestToJSONBasic(t *testing.T) {
	var buf bytes.Buffer
	if err := ToJSON(sampleFile(), &buf); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.Contains(s, `"id":1`) {
		t.Errorf("int inference missing: %q", s)
	}
	if !strings.Contains(s, `"active":true`) {
		t.Errorf("bool inference missing: %q", s)
	}
	if !strings.Contains(s, `"name":"Alice"`) {
		t.Errorf("string missing: %q", s)
	}
}

func TestToJSONHeaderOrderPreserved(t *testing.T) {
	var buf bytes.Buffer
	ToJSON(sampleFile(), &buf)
	s := buf.String()
	idPos := strings.Index(s, `"id"`)
	namePos := strings.Index(s, `"name"`)
	activePos := strings.Index(s, `"active"`)
	if idPos < 0 || !(idPos < namePos && namePos < activePos) {
		t.Errorf("order not preserved: %q", s)
	}
}

func TestToJSONWithoutTypeInfer(t *testing.T) {
	var buf bytes.Buffer
	ToJSON(sampleFile(), &buf, WithTypeInfer(false))
	s := buf.String()
	if !strings.Contains(s, `"id":"1"`) {
		t.Errorf("expected string id: %q", s)
	}
}

func TestToJSONNullForEmpty(t *testing.T) {
	f := gocsv.NewFile()
	f.SetHeaders([]string{"a", "b"})
	f.AppendStrRow([]string{"", "x"})
	var buf bytes.Buffer
	ToJSON(f, &buf, WithNullForEmpty(true))
	s := buf.String()
	if !strings.Contains(s, `"a":null`) {
		t.Errorf("expected null: %q", s)
	}
}

func TestToNDJSONBasic(t *testing.T) {
	var buf bytes.Buffer
	if err := ToNDJSON(sampleFile(), &buf); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %q", len(lines), buf.String())
	}
	for _, l := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Errorf("invalid json line: %q", l)
		}
	}
}

func TestFromJSON(t *testing.T) {
	data := `[{"id":1,"name":"Alice"},{"id":2,"name":"Bob"}]`
	f, err := FromJSON(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if f.RowCount() != 2 {
		t.Errorf("rows: %d", f.RowCount())
	}
	if got, _ := f.GetByHeader(0, "name"); got != "Alice" {
		t.Errorf("name: %q", got)
	}
}

func TestFromNDJSONStreaming(t *testing.T) {
	var buf bytes.Buffer
	for i := 0; i < 100; i++ {
		buf.WriteString(`{"id":`)
		buf.WriteString("1")
		buf.WriteString(`,"name":"x"}`)
		buf.WriteByte('\n')
	}
	f, err := FromNDJSON(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if f.RowCount() != 100 {
		t.Errorf("rows: %d", f.RowCount())
	}
}

func TestRoundTripJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := ToJSON(sampleFile(), &buf); err != nil {
		t.Fatal(err)
	}
	f, err := FromJSON(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if f.RowCount() != 2 {
		t.Errorf("rows: %d", f.RowCount())
	}
}

func TestBigIntPreserved(t *testing.T) {
	data := `[{"id":9007199254740993}]`
	f, err := FromJSON(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := f.GetByHeader(0, "id"); got != "9007199254740993" {
		t.Errorf("big int lost precision: %q", got)
	}
}

func TestDecoderStreaming(t *testing.T) {
	data := `{"a":1,"b":"x"}` + "\n" + `{"a":2,"b":"y"}` + "\n"
	d, err := NewDecoder(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	count := 0
	for d.Next() {
		count++
		rec := d.Record()
		if rec["a"] == "" || rec["b"] == "" {
			t.Errorf("empty record: %v", rec)
		}
	}
	if count != 2 {
		t.Errorf("count: %d", count)
	}
}

func TestEncoderStreaming(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf, []string{"x", "y"})
	if err := enc.Encode([]string{"1", "alpha"}); err != nil {
		t.Fatal(err)
	}
	if err := enc.Encode([]string{"2", "beta"}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("lines: %d", len(lines))
	}
	if !strings.Contains(lines[0], `"x":1`) || !strings.Contains(lines[0], `"y":"alpha"`) {
		t.Errorf("line 0: %q", lines[0])
	}
}

func TestStreamReaderAsRowIterator(t *testing.T) {
	data := `{"a":"1","b":"x"}` + "\n" + `{"a":"2","b":"y"}` + "\n"
	it, err := StreamReader(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer it.Close()
	count := 0
	for it.Next() {
		count++
	}
	if it.Error() != nil {
		t.Fatal(it.Error())
	}
	if count != 2 {
		t.Errorf("count: %d", count)
	}
}

func TestNestedObjectAsString(t *testing.T) {
	data := `[{"id":1,"addr":{"city":"Paris"}}]`
	f, err := FromJSON(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	got, _ := f.GetByHeader(0, "addr")
	if !strings.Contains(got, `"city":"Paris"`) {
		t.Errorf("expected stringified object, got: %q", got)
	}
}
