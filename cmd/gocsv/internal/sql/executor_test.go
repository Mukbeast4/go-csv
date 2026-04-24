package sql

import (
	"testing"

	gocsv "github.com/mukbeast4/go-csv"
)

func fixture() *gocsv.File {
	f := gocsv.NewFile()
	f.SetHeaders([]string{"id", "name", "age", "city"})
	f.AppendRow([]any{1, "Alice", 30, "Paris"})
	f.AppendRow([]any{2, "Bob", 25, "Berlin"})
	f.AppendRow([]any{3, "Charlie", 35, "Paris"})
	f.AppendRow([]any{4, "Diana", 28, "Madrid"})
	f.AppendRow([]any{5, "Eve", 42, "Berlin"})
	return f
}

func exec(t *testing.T, query string) *gocsv.File {
	t.Helper()
	stmt, err := Parse(query)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	f, err := Execute(stmt, fixture())
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	return f
}

func TestExecSelectAll(t *testing.T) {
	f := exec(t, "SELECT * FROM t")
	if f.RowCount() != 5 {
		t.Errorf("rows: %d", f.RowCount())
	}
}

func TestExecSelectCols(t *testing.T) {
	f := exec(t, "SELECT name, age FROM t")
	if len(f.Headers()) != 2 || f.ColCount() != 2 {
		t.Errorf("cols: %v", f.Headers())
	}
}

func TestExecWhere(t *testing.T) {
	f := exec(t, "SELECT * FROM t WHERE age >= 30")
	if f.RowCount() != 3 {
		t.Errorf("rows: %d", f.RowCount())
	}
}

func TestExecWhereAnd(t *testing.T) {
	f := exec(t, "SELECT * FROM t WHERE age >= 28 AND city = 'Paris'")
	if f.RowCount() != 2 {
		t.Errorf("rows: %d", f.RowCount())
	}
}

func TestExecWhereOr(t *testing.T) {
	f := exec(t, "SELECT * FROM t WHERE city = 'Paris' OR city = 'Madrid'")
	if f.RowCount() != 3 {
		t.Errorf("rows: %d", f.RowCount())
	}
}

func TestExecOrderBy(t *testing.T) {
	f := exec(t, "SELECT name FROM t ORDER BY age DESC LIMIT 1")
	rows, _ := f.GetRows()
	if rows[0][0] != "Eve" {
		t.Errorf("first: %v", rows[0])
	}
}

func TestExecLimit(t *testing.T) {
	f := exec(t, "SELECT * FROM t LIMIT 2")
	if f.RowCount() != 2 {
		t.Errorf("rows: %d", f.RowCount())
	}
}

func TestExecCount(t *testing.T) {
	f := exec(t, "SELECT COUNT(*) FROM t")
	rows, _ := f.GetRows()
	if rows[0][0] != "5" {
		t.Errorf("count: %v", rows[0])
	}
}

func TestExecAvg(t *testing.T) {
	f := exec(t, "SELECT AVG(age) FROM t")
	rows, _ := f.GetRows()
	expected := (30.0 + 25 + 35 + 28 + 42) / 5
	if rows[0][0] != formatFloat(expected) {
		t.Errorf("avg: %v (expected %v)", rows[0][0], formatFloat(expected))
	}
}

func TestExecGroupBy(t *testing.T) {
	f := exec(t, "SELECT city, COUNT(*) FROM t GROUP BY city")
	if f.RowCount() != 3 {
		t.Errorf("groups: %d", f.RowCount())
	}
	recs, _ := f.GetRecords()
	m := make(map[string]string)
	for _, r := range recs {
		m[r["city"]] = r["COUNT(*)"]
	}
	if m["Paris"] != "2" || m["Berlin"] != "2" || m["Madrid"] != "1" {
		t.Errorf("counts: %v", m)
	}
}

func TestExecGroupByMulti(t *testing.T) {
	f := exec(t, "SELECT city, SUM(age), AVG(age) FROM t GROUP BY city ORDER BY city ASC")
	if f.RowCount() != 3 {
		t.Errorf("groups: %d", f.RowCount())
	}
	headers := f.Headers()
	if headers[0] != "city" || headers[1] != "SUM(age)" || headers[2] != "AVG(age)" {
		t.Errorf("headers: %v", headers)
	}
}

func TestExecLike(t *testing.T) {
	f := exec(t, "SELECT name FROM t WHERE name LIKE 'A%'")
	if f.RowCount() != 1 {
		t.Errorf("rows: %d", f.RowCount())
	}
	rows, _ := f.GetRows()
	if rows[0][0] != "Alice" {
		t.Errorf("row: %v", rows[0])
	}
}

func TestExecNot(t *testing.T) {
	f := exec(t, "SELECT * FROM t WHERE NOT city = 'Paris'")
	if f.RowCount() != 3 {
		t.Errorf("rows: %d", f.RowCount())
	}
}

func TestExecAlias(t *testing.T) {
	f := exec(t, "SELECT age AS years FROM t LIMIT 1")
	headers := f.Headers()
	if headers[0] != "years" {
		t.Errorf("alias: %v", headers)
	}
}

func TestExecAggregateNoGroup(t *testing.T) {
	f := exec(t, "SELECT MIN(age), MAX(age) FROM t")
	rows, _ := f.GetRows()
	if rows[0][0] != "25" || rows[0][1] != "42" {
		t.Errorf("min/max: %v", rows[0])
	}
}

func TestExecIsNull(t *testing.T) {
	f := gocsv.NewFile()
	f.SetHeaders([]string{"id", "email"})
	f.AppendRow([]any{1, "a@a.com"})
	f.AppendRow([]any{2, ""})
	f.AppendRow([]any{3, "c@c.com"})
	stmt, _ := Parse("SELECT * FROM t WHERE email IS NULL")
	out, err := Execute(stmt, f)
	if err != nil {
		t.Fatal(err)
	}
	if out.RowCount() != 1 {
		t.Errorf("rows: %d", out.RowCount())
	}
}
