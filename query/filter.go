package query

func (q *Query) Where(pred func(Row) bool) *Query {
	if q.err != nil {
		return q
	}
	idx := q.headerIndex()
	out := make([][]string, 0, len(q.rows))
	for _, r := range q.rows {
		if pred(newRow(r, q.headers, idx)) {
			out = append(out, r)
		}
	}
	return &Query{headers: q.headers, rows: out}
}

func (q *Query) WhereNot(pred func(Row) bool) *Query {
	return q.Where(func(r Row) bool { return !pred(r) })
}

func (q *Query) WhereEq(col, value string) *Query {
	return q.Where(func(r Row) bool { return r.Get(col) == value })
}

func (q *Query) WhereIn(col string, values ...string) *Query {
	set := make(map[string]struct{}, len(values))
	for _, v := range values {
		set[v] = struct{}{}
	}
	return q.Where(func(r Row) bool {
		_, ok := set[r.Get(col)]
		return ok
	})
}

func (q *Query) WhereNotEmpty(col string) *Query {
	return q.Where(func(r Row) bool { return r.Get(col) != "" })
}
