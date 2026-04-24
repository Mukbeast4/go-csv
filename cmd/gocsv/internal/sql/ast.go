package sql

type Statement struct {
	Select   []Projection
	From     string
	Where    Expr
	GroupBy  []string
	OrderBy  []OrderClause
	Limit    int
	HasLimit bool
}

type Projection struct {
	Agg    string
	Column string
	Alias  string
	Star   bool
}

type OrderClause struct {
	Column string
	Desc   bool
}

type Expr interface {
	expr()
}

type BinaryExpr struct {
	Left  Expr
	Op    string
	Right Expr
}

func (*BinaryExpr) expr() {}

type ColumnRef struct {
	Name string
}

func (*ColumnRef) expr() {}

type Literal struct {
	Value    string
	IsNumber bool
}

func (*Literal) expr() {}

type NotExpr struct {
	Inner Expr
}

func (*NotExpr) expr() {}
