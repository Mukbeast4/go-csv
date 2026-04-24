package main

import (
	"fmt"
	"os"
	"strings"

	gocsv "github.com/mukbeast4/go-csv"
	"github.com/mukbeast4/go-csv/jsonx"
)

func main() {
	f := gocsv.NewFile()
	f.SetHeaders([]string{"id", "name", "active"})
	f.AppendStrRow([]string{"1", "Alice", "true"})
	f.AppendStrRow([]string{"2", "Bob", "false"})

	fmt.Println("=== JSON ===")
	jsonx.ToJSON(f, os.Stdout, jsonx.WithPretty(true))

	fmt.Println("\n=== NDJSON ===")
	jsonx.ToNDJSON(f, os.Stdout)

	fmt.Println("\n=== FromJSON round-trip ===")
	var buf strings.Builder
	jsonx.ToJSON(f, &buf)
	f2, err := jsonx.FromJSON(strings.NewReader(buf.String()))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Decoded %d rows, headers=%v\n", f2.RowCount(), f2.Headers())
}
