package sql

import "testing"

func TestParseSelectSimple(t *testing.T) {
	s, err := Parse("SELECT name, age FROM t")
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Select) != 2 || s.Select[0].Column != "name" || s.Select[1].Column != "age" {
		t.Errorf("projections: %+v", s.Select)
	}
	if s.From != "t" {
		t.Errorf("from: %q", s.From)
	}
}

func TestParseStar(t *testing.T) {
	s, err := Parse("SELECT * FROM t")
	if err != nil {
		t.Fatal(err)
	}
	if !s.Select[0].Star {
		t.Error("expected *")
	}
}

func TestParseAggregate(t *testing.T) {
	s, err := Parse("SELECT COUNT(*), AVG(age) FROM t")
	if err != nil {
		t.Fatal(err)
	}
	if s.Select[0].Agg != "COUNT" || s.Select[0].Column != "*" {
		t.Errorf("count: %+v", s.Select[0])
	}
	if s.Select[1].Agg != "AVG" || s.Select[1].Column != "age" {
		t.Errorf("avg: %+v", s.Select[1])
	}
}

func TestParseWhereBinary(t *testing.T) {
	s, err := Parse("SELECT * FROM t WHERE age > 30")
	if err != nil {
		t.Fatal(err)
	}
	bin, ok := s.Where.(*BinaryExpr)
	if !ok {
		t.Fatalf("expected BinaryExpr, got %T", s.Where)
	}
	if bin.Op != ">" {
		t.Errorf("op: %s", bin.Op)
	}
}

func TestParseWhereAndOr(t *testing.T) {
	s, err := Parse("SELECT * FROM t WHERE age > 30 AND city = 'Paris' OR role = 'admin'")
	if err != nil {
		t.Fatal(err)
	}
	or, ok := s.Where.(*BinaryExpr)
	if !ok || or.Op != "OR" {
		t.Fatalf("expected top OR, got %+v", s.Where)
	}
}

func TestParseGroupBy(t *testing.T) {
	s, err := Parse("SELECT city, COUNT(*) FROM t GROUP BY city")
	if err != nil {
		t.Fatal(err)
	}
	if len(s.GroupBy) != 1 || s.GroupBy[0] != "city" {
		t.Errorf("group by: %v", s.GroupBy)
	}
}

func TestParseOrderBy(t *testing.T) {
	s, err := Parse("SELECT * FROM t ORDER BY age DESC, name ASC")
	if err != nil {
		t.Fatal(err)
	}
	if len(s.OrderBy) != 2 {
		t.Fatalf("order by: %v", s.OrderBy)
	}
	if !s.OrderBy[0].Desc || s.OrderBy[1].Desc {
		t.Errorf("desc: %v", s.OrderBy)
	}
}

func TestParseLimit(t *testing.T) {
	s, err := Parse("SELECT * FROM t LIMIT 10")
	if err != nil {
		t.Fatal(err)
	}
	if !s.HasLimit || s.Limit != 10 {
		t.Errorf("limit: %d has=%v", s.Limit, s.HasLimit)
	}
}

func TestParseLike(t *testing.T) {
	s, err := Parse("SELECT * FROM t WHERE name LIKE 'A%'")
	if err != nil {
		t.Fatal(err)
	}
	bin := s.Where.(*BinaryExpr)
	if bin.Op != "LIKE" {
		t.Errorf("op: %s", bin.Op)
	}
}

func TestParseIsNull(t *testing.T) {
	s, err := Parse("SELECT * FROM t WHERE email IS NULL")
	if err != nil {
		t.Fatal(err)
	}
	bin := s.Where.(*BinaryExpr)
	if bin.Op != "IS" {
		t.Errorf("op: %s", bin.Op)
	}
}

func TestParseIsNotNull(t *testing.T) {
	s, err := Parse("SELECT * FROM t WHERE email IS NOT NULL")
	if err != nil {
		t.Fatal(err)
	}
	bin := s.Where.(*BinaryExpr)
	if bin.Op != "IS NOT" {
		t.Errorf("op: %s", bin.Op)
	}
}

func TestParseAlias(t *testing.T) {
	s, err := Parse("SELECT age AS years FROM t")
	if err != nil {
		t.Fatal(err)
	}
	if s.Select[0].Alias != "years" {
		t.Errorf("alias: %q", s.Select[0].Alias)
	}
}

func TestParseParentheses(t *testing.T) {
	s, err := Parse("SELECT * FROM t WHERE (age > 30 AND name = 'Alice') OR role = 'admin'")
	if err != nil {
		t.Fatal(err)
	}
	or := s.Where.(*BinaryExpr)
	if or.Op != "OR" {
		t.Errorf("op: %s", or.Op)
	}
}
