package marshal

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

type User struct {
	ID      int       `csv:"id"`
	Name    string    `csv:"name"`
	Email   string    `csv:"email,omitempty"`
	Age     int       `csv:"age"`
	Active  bool      `csv:"active"`
	Created time.Time `csv:"created,format=2006-01-02"`
	Tags    []string  `csv:"tags,sep=|"`
}

func sampleUsers() []User {
	return []User{
		{ID: 1, Name: "Alice", Email: "alice@a.com", Age: 30, Active: true,
			Created: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			Tags:    []string{"admin", "paris"}},
		{ID: 2, Name: "Bob", Age: 25, Active: false,
			Created: time.Date(2025, 3, 20, 0, 0, 0, 0, time.UTC),
			Tags:    []string{"user"}},
	}
}

func TestMarshalBasic(t *testing.T) {
	data, err := Marshal(sampleUsers())
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, "id,name,email,age,active,created,tags") {
		t.Errorf("headers missing: %q", s)
	}
	if !strings.Contains(s, "1,Alice,alice@a.com,30,true,2025-01-15,admin|paris") {
		t.Errorf("alice row missing: %q", s)
	}
}

func TestUnmarshalBasic(t *testing.T) {
	csv := "id,name,email,age,active,created,tags\n" +
		"1,Alice,alice@a.com,30,true,2025-01-15,admin|paris\n" +
		"2,Bob,,25,false,2025-03-20,user\n"
	var users []User
	if err := Unmarshal([]byte(csv), &users); err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 {
		t.Fatalf("count: %d", len(users))
	}
	if users[0].Name != "Alice" || users[0].Age != 30 {
		t.Errorf("alice: %+v", users[0])
	}
	if len(users[0].Tags) != 2 || users[0].Tags[0] != "admin" {
		t.Errorf("tags: %v", users[0].Tags)
	}
	if !users[0].Active {
		t.Errorf("active: %v", users[0].Active)
	}
}

func TestRoundTrip(t *testing.T) {
	original := sampleUsers()
	data, err := Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var decoded []User
	if err := Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded) != len(original) {
		t.Fatalf("count: %d", len(decoded))
	}
	for i := range original {
		if decoded[i].Name != original[i].Name || decoded[i].Age != original[i].Age {
			t.Errorf("row %d mismatch: %+v vs %+v", i, decoded[i], original[i])
		}
	}
}

func TestRequiredField(t *testing.T) {
	type R struct {
		ID   int    `csv:"id"`
		Name string `csv:"name,required"`
	}
	csv := "id,name\n1,\n"
	var rs []R
	err := Unmarshal([]byte(csv), &rs)
	if err == nil {
		t.Error("expected error for required empty field")
	}
}

func TestSkipField(t *testing.T) {
	type S struct {
		ID     int    `csv:"id"`
		Secret string `csv:"-"`
	}
	items := []S{{ID: 1, Secret: "xxx"}}
	data, _ := Marshal(items)
	if strings.Contains(string(data), "xxx") {
		t.Errorf("skipped field leaked: %q", data)
	}
}

func TestEncoderDecoder(t *testing.T) {
	var buf bytes.Buffer
	enc, err := NewEncoder(&buf, User{})
	if err != nil {
		t.Fatal(err)
	}
	for _, u := range sampleUsers() {
		if err := enc.Encode(u); err != nil {
			t.Fatal(err)
		}
	}
	enc.Close()

	dec, err := NewDecoder(&buf, User{})
	if err != nil {
		t.Fatal(err)
	}
	defer dec.Close()
	count := 0
	for {
		var u User
		err := dec.Decode(&u)
		if err != nil {
			break
		}
		count++
	}
	if count != 2 {
		t.Errorf("count: %d", count)
	}
}

func TestMarshalPointers(t *testing.T) {
	users := []*User{{ID: 1, Name: "Alice", Age: 30}}
	data, err := Marshal(users)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Alice") {
		t.Errorf("got %q", data)
	}
}

func TestUnmarshalPointers(t *testing.T) {
	csv := "id,name,email,age,active,created,tags\n1,Alice,,30,true,2025-01-15,\n"
	var users []*User
	if err := Unmarshal([]byte(csv), &users); err != nil {
		t.Fatal(err)
	}
	if len(users) != 1 || users[0].Name != "Alice" {
		t.Errorf("got %+v", users)
	}
}

func TestTagFormat(t *testing.T) {
	type E struct {
		When time.Time `csv:"when,format=02/01/2006"`
	}
	items := []E{{When: time.Date(2025, 4, 15, 0, 0, 0, 0, time.UTC)}}
	data, _ := Marshal(items)
	if !strings.Contains(string(data), "15/04/2025") {
		t.Errorf("format: %q", data)
	}
}

func TestOmitEmpty(t *testing.T) {
	csv := "id,name,email,age,active,created,tags\n1,Alice,,30,true,2025-01-15,\n"
	var users []User
	if err := Unmarshal([]byte(csv), &users); err != nil {
		t.Fatal(err)
	}
	if users[0].Email != "" {
		t.Errorf("email: %q", users[0].Email)
	}
}
