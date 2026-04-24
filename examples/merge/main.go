package main

import (
	"os"

	gocsv "github.com/mukbeast4/go-csv"
	"github.com/mukbeast4/go-csv/merge"
)

func main() {
	users := gocsv.NewFile()
	users.SetHeaders([]string{"id", "name"})
	users.AppendRow([]any{1, "Alice"})
	users.AppendRow([]any{2, "Bob"})
	users.AppendRow([]any{3, "Charlie"})

	orders := gocsv.NewFile()
	orders.SetHeaders([]string{"id", "amount"})
	orders.AppendRow([]any{1, 100})
	orders.AppendRow([]any{2, 200})
	orders.AppendRow([]any{4, 400})

	joined, _ := merge.LeftJoin(users, orders, merge.On("id"))
	joined.Write(os.Stdout)
}
