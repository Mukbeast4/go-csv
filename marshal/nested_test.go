package marshal

import (
	"strings"
	"testing"
)

type address struct {
	Street string `csv:"street"`
	City   string `csv:"city"`
}

type employee struct {
	ID      int     `csv:"id"`
	Name    string  `csv:"name"`
	Address address `csv:"addr,flatten"`
}

func TestFlattenBasic(t *testing.T) {
	rows := []employee{
		{ID: 1, Name: "Alice", Address: address{Street: "Rue 1", City: "Paris"}},
		{ID: 2, Name: "Bob", Address: address{Street: "Rue 2", City: "Lyon"}},
	}
	data, err := Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "id,name,addr.street,addr.city") {
		t.Errorf("headers: %q", s)
	}
	if !strings.Contains(s, "1,Alice,Rue 1,Paris") {
		t.Errorf("row: %q", s)
	}

	var got []employee
	if err := Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Address.City != "Paris" || got[1].Address.Street != "Rue 2" {
		t.Errorf("round-trip: %+v", got)
	}
}

type contact struct {
	Home address `csv:"home,flatten,prefix=home_"`
	Work address `csv:"work,flatten,prefix=work_"`
}

type person struct {
	Name    string  `csv:"name"`
	Contact contact `csv:"c,flatten"`
}

func TestFlattenNestedWithPrefix(t *testing.T) {
	rows := []person{
		{Name: "Alice", Contact: contact{
			Home: address{Street: "A", City: "Paris"},
			Work: address{Street: "B", City: "Lyon"},
		}},
	}
	data, err := Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	want := []string{"name", "home_street", "home_city", "work_street", "work_city"}
	for _, w := range want {
		if !strings.Contains(s, w) {
			t.Errorf("missing %q in %q", w, s)
		}
	}
}

type cyclic struct {
	Name string  `csv:"name"`
	Self *cyclic `csv:"self,flatten"`
}

func TestFlattenCycleDetection(t *testing.T) {
	_, err := Marshal([]cyclic{{Name: "a"}})
	if err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Errorf("expected cycle error, got: %v", err)
	}
}

type point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type jsonRow struct {
	ID    int   `csv:"id"`
	Coord point `csv:"coord,json"`
}

func TestJSONTagRoundTrip(t *testing.T) {
	rows := []jsonRow{
		{ID: 1, Coord: point{X: 10, Y: 20}},
		{ID: 2, Coord: point{X: -5, Y: 7}},
	}
	data, err := Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, `"{""x"":10,""y"":20}"`) {
		t.Errorf("json column not emitted: %q", s)
	}

	var got []jsonRow
	if err := Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].Coord.X != 10 || got[1].Coord.Y != 7 {
		t.Errorf("round-trip: %+v", got)
	}
}

type nonStruct struct {
	Name string `csv:"name"`
	Bad  int    `csv:"bad,flatten"`
}

func TestFlattenOnNonStructErrors(t *testing.T) {
	_, err := Marshal([]nonStruct{{Name: "x"}})
	if err == nil || !strings.Contains(err.Error(), "flatten requires struct") {
		t.Errorf("expected flatten-type error, got: %v", err)
	}
}
