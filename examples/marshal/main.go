package main

import (
	"fmt"
	"time"

	"github.com/mukbeast4/go-csv/marshal"
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

func main() {
	users := []User{
		{ID: 1, Name: "Alice", Email: "alice@a.com", Age: 30, Active: true,
			Created: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			Tags:    []string{"admin", "paris"}},
		{ID: 2, Name: "Bob", Age: 25, Active: false,
			Created: time.Date(2025, 3, 20, 0, 0, 0, 0, time.UTC),
			Tags:    []string{"user"}},
	}

	data, err := marshal.Marshal(users)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))

	var decoded []User
	if err := marshal.Unmarshal(data, &decoded); err != nil {
		panic(err)
	}
	for _, u := range decoded {
		fmt.Printf("Decoded: %+v\n", u)
	}
}
