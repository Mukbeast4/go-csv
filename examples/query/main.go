package main

import (
	"fmt"

	gocsv "github.com/mukbeast4/go-csv"
	"github.com/mukbeast4/go-csv/query"
)

func main() {
	f := gocsv.NewFile()
	f.SetHeaders([]string{"name", "age", "city"})
	f.AppendRow([]any{"Alice", 30, "Paris"})
	f.AppendRow([]any{"Bob", 25, "Berlin"})
	f.AppendRow([]any{"Charlie", 35, "Paris"})
	f.AppendRow([]any{"Diana", 28, "Madrid"})

	q := query.From(f).
		Where(func(r query.Row) bool { return r.Int("age") >= 28 }).
		OrderBy("age", query.Desc).
		Limit(3)

	fmt.Printf("Count: %d\n", q.Count())
	fmt.Printf("Avg age: %.1f\n", q.Avg("age"))
	fmt.Printf("Cities: %v\n", q.CountBy("city"))

	q.ForEach(func(r query.Row) {
		fmt.Printf("  %s (%d) — %s\n", r.Get("name"), r.Int("age"), r.Get("city"))
	})
}
