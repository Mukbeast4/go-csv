package main

import (
	"fmt"
	"os"
	"path/filepath"

	gocsv "github.com/mukbeast4/go-csv"
	"github.com/mukbeast4/go-csv/compress"
)

func main() {
	dir, _ := os.MkdirTemp("", "gocsv-compress-*")
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "data.csv.gz")

	f := gocsv.NewFile()
	f.SetHeaders([]string{"id", "name"})
	for i := 1; i <= 1000; i++ {
		f.AppendRow([]any{i, fmt.Sprintf("user%d", i)})
	}
	if err := compress.SaveAs(f, path); err != nil {
		panic(err)
	}

	info, _ := os.Stat(path)
	fmt.Printf("Compressed file: %d bytes\n", info.Size())

	opened, err := compress.Open(path, gocsv.WithHeader(true))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Read back: %d rows\n", opened.RowCount())
}
