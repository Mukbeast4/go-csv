package query

func (q *Query) Select(cols ...string) *Query {
	if q.err != nil {
		return q
	}
	srcIdx := q.headerIndex()
	indexes := make([]int, 0, len(cols))
	for _, c := range cols {
		i, ok := srcIdx[c]
		if !ok {
			return &Query{err: &MissingColumnError{Column: c}}
		}
		indexes = append(indexes, i)
	}
	newHeaders := make([]string, len(cols))
	copy(newHeaders, cols)
	newRows := make([][]string, len(q.rows))
	for i, r := range q.rows {
		row := make([]string, len(indexes))
		for j, idx := range indexes {
			if idx < len(r) {
				row[j] = r[idx]
			}
		}
		newRows[i] = row
	}
	return &Query{headers: newHeaders, rows: newRows}
}

func (q *Query) Map(fn func(Row) []string) *Query {
	if q.err != nil {
		return q
	}
	idx := q.headerIndex()
	out := make([][]string, len(q.rows))
	for i, r := range q.rows {
		out[i] = fn(newRow(r, q.headers, idx))
	}
	return &Query{headers: q.headers, rows: out}
}

func (q *Query) Distinct(cols ...string) *Query {
	if q.err != nil {
		return q
	}
	var keyFn func([]string) string
	if len(cols) == 0 {
		keyFn = func(r []string) string { return joinFields(r) }
	} else {
		srcIdx := q.headerIndex()
		ids := make([]int, 0, len(cols))
		for _, c := range cols {
			i, ok := srcIdx[c]
			if !ok {
				return &Query{err: &MissingColumnError{Column: c}}
			}
			ids = append(ids, i)
		}
		keyFn = func(r []string) string {
			parts := make([]string, len(ids))
			for j, i := range ids {
				if i < len(r) {
					parts[j] = r[i]
				}
			}
			return joinFields(parts)
		}
	}
	seen := make(map[string]struct{})
	out := make([][]string, 0, len(q.rows))
	for _, r := range q.rows {
		k := keyFn(r)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, r)
	}
	return &Query{headers: q.headers, rows: out}
}

func (q *Query) WithColumn(name string, fn func(Row) string) *Query {
	if q.err != nil {
		return q
	}
	newHeaders := append([]string{}, q.headers...)
	newHeaders = append(newHeaders, name)
	idx := q.headerIndex()
	newRows := make([][]string, len(q.rows))
	for i, r := range q.rows {
		row := append([]string{}, r...)
		row = append(row, fn(newRow(r, q.headers, idx)))
		newRows[i] = row
	}
	return &Query{headers: newHeaders, rows: newRows}
}

func (q *Query) Rename(old, new string) *Query {
	if q.err != nil {
		return q
	}
	newHeaders := make([]string, len(q.headers))
	copy(newHeaders, q.headers)
	for i, h := range newHeaders {
		if h == old {
			newHeaders[i] = new
		}
	}
	return &Query{headers: newHeaders, rows: q.rows}
}

func joinFields(fields []string) string {
	total := len(fields)
	for _, f := range fields {
		total += len(f)
	}
	buf := make([]byte, 0, total)
	for i, f := range fields {
		if i > 0 {
			buf = append(buf, 0x1f)
		}
		buf = append(buf, f...)
	}
	return string(buf)
}
