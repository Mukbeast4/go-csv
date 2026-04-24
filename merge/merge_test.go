package merge

import (
	"testing"

	gocsv "github.com/mukbeast4/go-csv"
)

func usersFile() *gocsv.File {
	f := gocsv.NewFile()
	f.SetHeaders([]string{"id", "name"})
	f.AppendRow([]any{1, "Alice"})
	f.AppendRow([]any{2, "Bob"})
	f.AppendRow([]any{3, "Charlie"})
	return f
}

func ordersFile() *gocsv.File {
	f := gocsv.NewFile()
	f.SetHeaders([]string{"id", "amount"})
	f.AppendRow([]any{1, 100})
	f.AppendRow([]any{2, 200})
	f.AppendRow([]any{4, 400})
	return f
}

func TestInnerJoin(t *testing.T) {
	out, err := InnerJoin(usersFile(), ordersFile(), On("id"))
	if err != nil {
		t.Fatal(err)
	}
	if out.RowCount() != 2 {
		t.Errorf("rows: %d", out.RowCount())
	}
	headers := out.Headers()
	if len(headers) != 3 {
		t.Errorf("headers: %v", headers)
	}
}

func TestLeftJoin(t *testing.T) {
	out, err := LeftJoin(usersFile(), ordersFile(), On("id"))
	if err != nil {
		t.Fatal(err)
	}
	if out.RowCount() != 3 {
		t.Errorf("rows: %d", out.RowCount())
	}
}

func TestRightJoin(t *testing.T) {
	out, err := RightJoin(usersFile(), ordersFile(), On("id"))
	if err != nil {
		t.Fatal(err)
	}
	if out.RowCount() != 3 {
		t.Errorf("rows: %d", out.RowCount())
	}
}

func TestFullJoin(t *testing.T) {
	out, err := FullJoin(usersFile(), ordersFile(), On("id"))
	if err != nil {
		t.Fatal(err)
	}
	if out.RowCount() != 4 {
		t.Errorf("rows: %d", out.RowCount())
	}
}

func TestConcat(t *testing.T) {
	a := gocsv.NewFile()
	a.SetHeaders([]string{"id", "name"})
	a.AppendRow([]any{1, "Alice"})

	b := gocsv.NewFile()
	b.SetHeaders([]string{"id", "name"})
	b.AppendRow([]any{2, "Bob"})

	out, err := Concat(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if out.RowCount() != 2 {
		t.Errorf("rows: %d", out.RowCount())
	}
}

func TestUnionBy(t *testing.T) {
	a := gocsv.NewFile()
	a.SetHeaders([]string{"id", "name"})
	a.AppendRow([]any{1, "Alice"})
	a.AppendRow([]any{2, "Bob"})

	b := gocsv.NewFile()
	b.SetHeaders([]string{"id", "name"})
	b.AppendRow([]any{2, "Bob2"})
	b.AppendRow([]any{3, "Charlie"})

	out, err := UnionBy(a, b, On("id"))
	if err != nil {
		t.Fatal(err)
	}
	if out.RowCount() != 3 {
		t.Errorf("rows: %d", out.RowCount())
	}
}

func TestDiff(t *testing.T) {
	before := gocsv.NewFile()
	before.SetHeaders([]string{"id", "name"})
	before.AppendRow([]any{1, "Alice"})
	before.AppendRow([]any{2, "Bob"})
	before.AppendRow([]any{3, "Charlie"})

	after := gocsv.NewFile()
	after.SetHeaders([]string{"id", "name"})
	after.AppendRow([]any{1, "Alice"})
	after.AppendRow([]any{2, "Bobby"})
	after.AppendRow([]any{4, "Diana"})

	d, err := Diff(before, after, On("id"))
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Added) != 1 || d.Added[0][1] != "Diana" {
		t.Errorf("added: %v", d.Added)
	}
	if len(d.Removed) != 1 || d.Removed[0][1] != "Charlie" {
		t.Errorf("removed: %v", d.Removed)
	}
	if len(d.Modified) != 1 || d.Modified[0].After[1] != "Bobby" {
		t.Errorf("modified: %v", d.Modified)
	}
}

func TestOnComposite(t *testing.T) {
	a := gocsv.NewFile()
	a.SetHeaders([]string{"k1", "k2", "val"})
	a.AppendRow([]any{1, "x", "a"})

	b := gocsv.NewFile()
	b.SetHeaders([]string{"k1", "k2", "other"})
	b.AppendRow([]any{1, "x", "b"})
	b.AppendRow([]any{1, "y", "c"})

	out, err := InnerJoin(a, b, OnComposite("k1", "k2"))
	if err != nil {
		t.Fatal(err)
	}
	if out.RowCount() != 1 {
		t.Errorf("rows: %d", out.RowCount())
	}
}
