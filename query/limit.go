package query

func (q *Query) Limit(n int) *Query {
	if q.err != nil {
		return q
	}
	if n < 0 {
		n = 0
	}
	if n >= len(q.rows) {
		return q
	}
	rows := make([][]string, n)
	copy(rows, q.rows[:n])
	return &Query{headers: q.headers, rows: rows}
}

func (q *Query) Skip(n int) *Query {
	if q.err != nil {
		return q
	}
	if n < 0 {
		n = 0
	}
	if n >= len(q.rows) {
		return &Query{headers: q.headers, rows: nil}
	}
	rows := make([][]string, len(q.rows)-n)
	copy(rows, q.rows[n:])
	return &Query{headers: q.headers, rows: rows}
}

func (q *Query) First() (Row, bool) {
	if q.err != nil || len(q.rows) == 0 {
		return Row{}, false
	}
	return newRow(q.rows[0], q.headers, q.headerIndex()), true
}

func (q *Query) Last() (Row, bool) {
	if q.err != nil || len(q.rows) == 0 {
		return Row{}, false
	}
	return newRow(q.rows[len(q.rows)-1], q.headers, q.headerIndex()), true
}

func (q *Query) Reverse() *Query {
	if q.err != nil {
		return q
	}
	rows := make([][]string, len(q.rows))
	for i, r := range q.rows {
		rows[len(q.rows)-1-i] = r
	}
	return &Query{headers: q.headers, rows: rows}
}
