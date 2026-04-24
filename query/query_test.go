package query

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

func TestFromFile(t *testing.T) {
	q := From(fixture())
	if q.Count() != 5 {
		t.Errorf("count: %d", q.Count())
	}
	if len(q.Headers()) != 4 {
		t.Errorf("headers: %v", q.Headers())
	}
}

func TestFromRows(t *testing.T) {
	q := FromRows([][]string{{"1", "a"}, {"2", "b"}}, []string{"id", "val"})
	if q.Count() != 2 {
		t.Errorf("count: %d", q.Count())
	}
}

func TestWhere(t *testing.T) {
	q := From(fixture()).Where(func(r Row) bool {
		return r.Int("age") >= 30
	})
	if q.Count() != 3 {
		t.Errorf("count: %d", q.Count())
	}
}

func TestWhereEq(t *testing.T) {
	q := From(fixture()).WhereEq("city", "Paris")
	if q.Count() != 2 {
		t.Errorf("count: %d", q.Count())
	}
}

func TestWhereIn(t *testing.T) {
	q := From(fixture()).WhereIn("city", "Paris", "Madrid")
	if q.Count() != 3 {
		t.Errorf("count: %d", q.Count())
	}
}

func TestSelect(t *testing.T) {
	q := From(fixture()).Select("name", "age")
	headers := q.Headers()
	if len(headers) != 2 || headers[0] != "name" || headers[1] != "age" {
		t.Errorf("headers: %v", headers)
	}
	rows := q.ToRows()
	if rows[0][0] != "Alice" || rows[0][1] != "30" {
		t.Errorf("row: %v", rows[0])
	}
}

func TestSelectMissingColumn(t *testing.T) {
	q := From(fixture()).Select("missing")
	if q.Err() == nil {
		t.Error("expected error")
	}
}

func TestOrderByAsc(t *testing.T) {
	q := From(fixture()).OrderBy("age", Asc)
	rec, _ := q.First()
	if rec.Get("name") != "Bob" {
		t.Errorf("first: %s", rec.Get("name"))
	}
}

func TestOrderByDesc(t *testing.T) {
	q := From(fixture()).OrderBy("age", Desc)
	rec, _ := q.First()
	if rec.Get("name") != "Eve" {
		t.Errorf("first: %s", rec.Get("name"))
	}
}

func TestOrderByFn(t *testing.T) {
	q := From(fixture()).OrderByFn(func(a, b Row) bool {
		return a.Str("name") < b.Str("name")
	})
	rec, _ := q.First()
	if rec.Get("name") != "Alice" {
		t.Errorf("first: %s", rec.Get("name"))
	}
}

func TestDistinct(t *testing.T) {
	q := From(fixture()).Distinct("city")
	if q.Count() != 3 {
		t.Errorf("count: %d", q.Count())
	}
}

func TestLimit(t *testing.T) {
	q := From(fixture()).Limit(2)
	if q.Count() != 2 {
		t.Errorf("count: %d", q.Count())
	}
}

func TestSkip(t *testing.T) {
	q := From(fixture()).Skip(2)
	if q.Count() != 3 {
		t.Errorf("count: %d", q.Count())
	}
}

func TestCount(t *testing.T) {
	if From(fixture()).Count() != 5 {
		t.Error("count")
	}
}

func TestSum(t *testing.T) {
	s := From(fixture()).Sum("age")
	if s != 30+25+35+28+42 {
		t.Errorf("sum: %v", s)
	}
}

func TestAvg(t *testing.T) {
	avg := From(fixture()).Avg("age")
	if avg != (30+25+35+28+42)/5.0 {
		t.Errorf("avg: %v", avg)
	}
}

func TestMinMax(t *testing.T) {
	q := From(fixture())
	if q.Min("age") != 25 {
		t.Errorf("min: %v", q.Min("age"))
	}
	if q.Max("age") != 42 {
		t.Errorf("max: %v", q.Max("age"))
	}
}

func TestCountBy(t *testing.T) {
	m := From(fixture()).CountBy("city")
	if m["Paris"] != 2 || m["Berlin"] != 2 || m["Madrid"] != 1 {
		t.Errorf("counts: %v", m)
	}
}

func TestGroupBy(t *testing.T) {
	g := From(fixture()).GroupBy("city")
	if g["Paris"].Count() != 2 {
		t.Errorf("Paris: %d", g["Paris"].Count())
	}
}

func TestChaining(t *testing.T) {
	q := From(fixture()).
		Where(func(r Row) bool { return r.Int("age") >= 28 }).
		Select("name", "city").
		OrderBy("name", Asc).
		Limit(3)
	if q.Count() != 3 {
		t.Errorf("count: %d", q.Count())
	}
	r, _ := q.First()
	if r.Get("name") != "Alice" {
		t.Errorf("first: %s", r.Get("name"))
	}
}

func TestToFile(t *testing.T) {
	f := From(fixture()).Where(func(r Row) bool {
		return r.Int("age") > 30
	}).ToFile()
	if f.RowCount() != 2 {
		t.Errorf("rows: %d", f.RowCount())
	}
}

func TestWithColumn(t *testing.T) {
	q := From(fixture()).WithColumn("age_plus", func(r Row) string {
		return "X"
	})
	headers := q.Headers()
	if headers[len(headers)-1] != "age_plus" {
		t.Errorf("headers: %v", headers)
	}
	r, _ := q.First()
	if r.Get("age_plus") != "X" {
		t.Errorf("col: %s", r.Get("age_plus"))
	}
}

func TestRename(t *testing.T) {
	q := From(fixture()).Rename("age", "years")
	r, _ := q.First()
	if r.Get("years") != "30" {
		t.Errorf("rename: %s", r.Get("years"))
	}
}

func TestReverse(t *testing.T) {
	q := From(fixture()).Reverse()
	r, _ := q.First()
	if r.Get("name") != "Eve" {
		t.Errorf("reverse: %s", r.Get("name"))
	}
}

func TestForEach(t *testing.T) {
	count := 0
	From(fixture()).ForEach(func(r Row) { count++ })
	if count != 5 {
		t.Errorf("count: %d", count)
	}
}

func TestReduce(t *testing.T) {
	total := From(fixture()).Reduce(int64(0), func(acc any, r Row) any {
		return acc.(int64) + r.Int("age")
	}).(int64)
	if total != 30+25+35+28+42 {
		t.Errorf("total: %d", total)
	}
}
