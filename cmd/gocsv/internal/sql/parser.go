package sql

import (
	"fmt"
	"strconv"
	"strings"
)

type parser struct {
	toks []token
	pos  int
}

func Parse(query string) (*Statement, error) {
	toks, err := tokenize(query)
	if err != nil {
		return nil, err
	}
	p := &parser{toks: toks}
	return p.parseStatement()
}

func (p *parser) peek() token {
	if p.pos >= len(p.toks) {
		return token{kind: tokEOF}
	}
	return p.toks[p.pos]
}

func (p *parser) next() token {
	t := p.peek()
	p.pos++
	return t
}

func (p *parser) expectKeyword(k string) error {
	t := p.next()
	if t.kind != tokKeyword || t.val != k {
		return fmt.Errorf("expected %s, got %q", k, t.val)
	}
	return nil
}

func (p *parser) parseStatement() (*Statement, error) {
	if err := p.expectKeyword("SELECT"); err != nil {
		return nil, err
	}
	stmt := &Statement{}
	projs, err := p.parseProjections()
	if err != nil {
		return nil, err
	}
	stmt.Select = projs

	t := p.peek()
	if t.kind == tokKeyword && t.val == "FROM" {
		p.next()
		tn := p.next()
		if tn.kind != tokIdent {
			return nil, fmt.Errorf("expected table name, got %q", tn.val)
		}
		stmt.From = tn.val
	}

	if p.peek().kind == tokKeyword && p.peek().val == "WHERE" {
		p.next()
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		stmt.Where = expr
	}

	if p.peek().kind == tokKeyword && p.peek().val == "GROUP" {
		p.next()
		if err := p.expectKeyword("BY"); err != nil {
			return nil, err
		}
		cols, err := p.parseIdentList()
		if err != nil {
			return nil, err
		}
		stmt.GroupBy = cols
	}

	if p.peek().kind == tokKeyword && p.peek().val == "ORDER" {
		p.next()
		if err := p.expectKeyword("BY"); err != nil {
			return nil, err
		}
		for {
			col := p.next()
			if col.kind != tokIdent {
				return nil, fmt.Errorf("expected column name, got %q", col.val)
			}
			clause := OrderClause{Column: col.val}
			if p.peek().kind == tokKeyword {
				switch p.peek().val {
				case "ASC":
					p.next()
				case "DESC":
					p.next()
					clause.Desc = true
				}
			}
			stmt.OrderBy = append(stmt.OrderBy, clause)
			if p.peek().kind != tokComma {
				break
			}
			p.next()
		}
	}

	if p.peek().kind == tokKeyword && p.peek().val == "LIMIT" {
		p.next()
		n := p.next()
		if n.kind != tokNumber {
			return nil, fmt.Errorf("expected number after LIMIT, got %q", n.val)
		}
		v, err := strconv.Atoi(n.val)
		if err != nil {
			return nil, err
		}
		stmt.Limit = v
		stmt.HasLimit = true
	}

	if p.peek().kind != tokEOF {
		return nil, fmt.Errorf("unexpected token %q", p.peek().val)
	}
	return stmt, nil
}

func (p *parser) parseProjections() ([]Projection, error) {
	var projs []Projection
	for {
		proj, err := p.parseProjection()
		if err != nil {
			return nil, err
		}
		projs = append(projs, proj)
		if p.peek().kind != tokComma {
			break
		}
		p.next()
	}
	return projs, nil
}

func (p *parser) parseProjection() (Projection, error) {
	t := p.peek()
	if t.kind == tokStar {
		p.next()
		return Projection{Star: true}, nil
	}
	if t.kind == tokKeyword && isAggKeyword(t.val) {
		p.next()
		if p.peek().kind != tokLParen {
			return Projection{}, fmt.Errorf("expected ( after %s", t.val)
		}
		p.next()
		proj := Projection{Agg: t.val}
		inner := p.peek()
		if inner.kind == tokStar {
			proj.Column = "*"
			p.next()
		} else {
			col := p.next()
			if col.kind != tokIdent {
				return Projection{}, fmt.Errorf("expected column in %s(), got %q", t.val, col.val)
			}
			proj.Column = col.val
		}
		if p.peek().kind != tokRParen {
			return Projection{}, fmt.Errorf("expected )")
		}
		p.next()
		if p.peek().kind == tokKeyword && p.peek().val == "AS" {
			p.next()
			a := p.next()
			if a.kind != tokIdent {
				return Projection{}, fmt.Errorf("expected alias")
			}
			proj.Alias = a.val
		}
		return proj, nil
	}
	if t.kind == tokIdent {
		p.next()
		proj := Projection{Column: t.val}
		if p.peek().kind == tokKeyword && p.peek().val == "AS" {
			p.next()
			a := p.next()
			if a.kind != tokIdent {
				return Projection{}, fmt.Errorf("expected alias")
			}
			proj.Alias = a.val
		}
		return proj, nil
	}
	return Projection{}, fmt.Errorf("expected column or aggregate, got %q", t.val)
}

func (p *parser) parseIdentList() ([]string, error) {
	var out []string
	for {
		t := p.next()
		if t.kind != tokIdent {
			return nil, fmt.Errorf("expected identifier, got %q", t.val)
		}
		out = append(out, t.val)
		if p.peek().kind != tokComma {
			break
		}
		p.next()
	}
	return out, nil
}

func (p *parser) parseExpr() (Expr, error) {
	return p.parseOr()
}

func (p *parser) parseOr() (Expr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.peek().kind == tokKeyword && p.peek().val == "OR" {
		p.next()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Op: "OR", Right: right}
	}
	return left, nil
}

func (p *parser) parseAnd() (Expr, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}
	for p.peek().kind == tokKeyword && p.peek().val == "AND" {
		p.next()
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Op: "AND", Right: right}
	}
	return left, nil
}

func (p *parser) parseNot() (Expr, error) {
	if p.peek().kind == tokKeyword && p.peek().val == "NOT" {
		p.next()
		inner, err := p.parseCompare()
		if err != nil {
			return nil, err
		}
		return &NotExpr{Inner: inner}, nil
	}
	return p.parseCompare()
}

func (p *parser) parseCompare() (Expr, error) {
	if p.peek().kind == tokLParen {
		p.next()
		e, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		if p.peek().kind != tokRParen {
			return nil, fmt.Errorf("expected )")
		}
		p.next()
		return e, nil
	}
	left, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	tok := p.peek()
	if tok.kind == tokOp {
		p.next()
		right, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		return &BinaryExpr{Left: left, Op: tok.val, Right: right}, nil
	}
	if tok.kind == tokKeyword && tok.val == "LIKE" {
		p.next()
		right, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		return &BinaryExpr{Left: left, Op: "LIKE", Right: right}, nil
	}
	if tok.kind == tokKeyword && tok.val == "IS" {
		p.next()
		negate := false
		if p.peek().kind == tokKeyword && p.peek().val == "NOT" {
			p.next()
			negate = true
		}
		if p.peek().kind != tokKeyword || p.peek().val != "NULL" {
			return nil, fmt.Errorf("expected NULL")
		}
		p.next()
		op := "IS"
		if negate {
			op = "IS NOT"
		}
		return &BinaryExpr{Left: left, Op: op, Right: &Literal{Value: ""}}, nil
	}
	return left, nil
}

func (p *parser) parseValue() (Expr, error) {
	t := p.next()
	switch t.kind {
	case tokIdent:
		return &ColumnRef{Name: t.val}, nil
	case tokNumber:
		return &Literal{Value: t.val, IsNumber: true}, nil
	case tokString:
		return &Literal{Value: t.val}, nil
	case tokKeyword:
		if t.val == "NULL" {
			return &Literal{Value: ""}, nil
		}
	}
	return nil, fmt.Errorf("expected value, got %q", t.val)
}

func isAggKeyword(s string) bool {
	switch strings.ToUpper(s) {
	case "COUNT", "SUM", "AVG", "MIN", "MAX":
		return true
	}
	return false
}
