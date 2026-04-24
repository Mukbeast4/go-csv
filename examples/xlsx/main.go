package main

import (
	"fmt"
	"os"
	"path/filepath"

	gocsv "github.com/mukbeast4/go-csv"
	"github.com/mukbeast4/go-csv/xlsx"
)

func main() {
	dir, _ := os.MkdirTemp("", "gocsv-xlsx-*")
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "users.xlsx")

	f := gocsv.NewFile()
	f.SetHeaders([]string{"id", "name", "active"})
	f.AppendStrRow([]string{"1", "Alice", "true"})
	f.AppendStrRow([]string{"2", "Bob", "false"})

	if err := xlsx.ToXLSX(f, path, xlsx.WithSheetName("Users"), xlsx.WithAutoFilter(true)); err != nil {
		panic(err)
	}
	info, _ := os.Stat(path)
	fmt.Printf("Wrote %s (%d bytes)\n", path, info.Size())

	names, _ := xlsx.SheetNames(path)
	fmt.Printf("Sheets: %v\n", names)

	back, err := xlsx.FromXLSX(path, "Users")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Read back %d rows\n", back.RowCount())
}
