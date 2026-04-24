package marshal

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

type flexible struct {
	ID    int               `csv:"id"`
	Name  string            `csv:"name"`
	Extra map[string]string `csv:",rest"`
}

func TestRestDecodeBasic(t *testing.T) {
	csv := "id,name,dept,country\n1,Alice,eng,FR\n2,Bob,sales,UK\n"
	var rows []flexible
	if err := Unmarshal([]byte(csv), &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("count: %d", len(rows))
	}
	if rows[0].Name != "Alice" || rows[0].ID != 1 {
		t.Errorf("struct fields: %+v", rows[0])
	}
	if rows[0].Extra["dept"] != "eng" || rows[0].Extra["country"] != "FR" {
		t.Errorf("extra: %v", rows[0].Extra)
	}
	if rows[1].Extra["dept"] != "sales" {
		t.Errorf("bob extra: %v", rows[1].Extra)
	}
}

func TestRestEncodeUnionKeys(t *testing.T) {
	rows := []flexible{
		{ID: 1, Name: "Alice", Extra: map[string]string{"dept": "eng", "country": "FR"}},
		{ID: 2, Name: "Bob", Extra: map[string]string{"dept": "sales", "level": "senior"}},
	}
	data, err := Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "id,name,country,dept,level") {
		t.Errorf("header missing or unsorted: %q", s)
	}
	if !strings.Contains(s, "1,Alice,FR,eng,") {
		t.Errorf("row 1: %q", s)
	}
	if !strings.Contains(s, "2,Bob,,sales,senior") {
		t.Errorf("row 2: %q", s)
	}
}

func TestRestRoundTrip(t *testing.T) {
	original := []flexible{
		{ID: 1, Name: "Alice", Extra: map[string]string{"dept": "eng", "country": "FR"}},
		{ID: 2, Name: "Bob", Extra: map[string]string{"dept": "sales", "country": "UK"}},
	}
	data, err := Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var back []flexible
	if err := Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(original, back) {
		t.Errorf("round-trip mismatch:\n  original=%+v\n  got=%+v", original, back)
	}
}

type twoRest struct {
	A map[string]string `csv:",rest"`
	B map[string]string `csv:",rest"`
}

func TestRestMultipleFieldsErrors(t *testing.T) {
	// clear cached structInfo since the type was never used before; still safe.
	_, err := Marshal([]twoRest{})
	if err == nil || !strings.Contains(err.Error(), "multiple rest") {
		t.Errorf("want multiple rest error, got: %v", err)
	}
}

type badRestType struct {
	Extra map[string]int `csv:",rest"`
}

func TestRestWrongMapValueType(t *testing.T) {
	_, err := Marshal([]badRestType{})
	if err == nil || !strings.Contains(err.Error(), "map[string]string") {
		t.Errorf("want wrong-type error, got: %v", err)
	}
}

type collisionRest struct {
	Name  string            `csv:"name"`
	Extra map[string]string `csv:",rest"`
}

func TestRestKeyCollisionWithStructHeader(t *testing.T) {
	rows := []collisionRest{{Name: "x", Extra: map[string]string{"name": "boom"}}}
	_, err := Marshal(rows)
	if err == nil || !strings.Contains(err.Error(), "collides") {
		t.Errorf("want collision error, got: %v", err)
	}
}

func TestRestEncoderRequiresSetRestKeys(t *testing.T) {
	var buf bytes.Buffer
	enc, err := NewEncoder(&buf, flexible{})
	if err != nil {
		t.Fatal(err)
	}
	defer enc.Close()
	err = enc.Encode(flexible{ID: 1, Name: "x", Extra: map[string]string{"a": "1"}})
	if err == nil || !strings.Contains(err.Error(), "SetRestKeys") {
		t.Errorf("want SetRestKeys error, got: %v", err)
	}
}

func TestRestEncoderWithSetRestKeys(t *testing.T) {
	var buf bytes.Buffer
	enc, err := NewEncoder(&buf, flexible{})
	if err != nil {
		t.Fatal(err)
	}
	enc.SetRestKeys([]string{"country", "dept"})
	if err := enc.Encode(flexible{ID: 1, Name: "Alice", Extra: map[string]string{"dept": "eng", "country": "FR"}}); err != nil {
		t.Fatal(err)
	}
	if err := enc.Close(); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.Contains(s, "id,name,country,dept") || !strings.Contains(s, "1,Alice,FR,eng") {
		t.Errorf("bad output: %q", s)
	}
}
