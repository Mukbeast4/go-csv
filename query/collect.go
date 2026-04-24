package query

import (
	gocsv "github.com/mukbeast4/go-csv"
)

func (q *Query) ToRows() [][]string {
	if q.err != nil {
		return nil
	}
	out := make([][]string, len(q.rows))
	for i, r := range q.rows {
		cp := make([]string, len(r))
		copy(cp, r)
		out[i] = cp
	}
	return out
}

func (q *Query) ToRecords() []map[string]string {
	if q.err != nil {
		return nil
	}
	out := make([]map[string]string, len(q.rows))
	for i, r := range q.rows {
		m := make(map[string]string, len(q.headers))
		for j, h := range q.headers {
			if j < len(r) {
				m[h] = r[j]
			} else {
				m[h] = ""
			}
		}
		out[i] = m
	}
	return out
}

func (q *Query) ToFile(opts ...gocsv.Option) *gocsv.File {
	if q.err != nil {
		return nil
	}
	f := gocsv.NewFile(opts...)
	if len(q.headers) > 0 {
		f.SetHeaders(q.headers)
	}
	for _, r := range q.rows {
		f.AppendStrRow(r)
	}
	return f
}

func (q *Query) SaveAs(path string, opts ...gocsv.Option) error {
	f := q.ToFile(opts...)
	if f == nil {
		return q.err
	}
	return f.SaveAs(path)
}

func (q *Query) ForEach(fn func(Row)) {
	if q.err != nil {
		return
	}
	idx := q.headerIndex()
	for _, r := range q.rows {
		fn(newRow(r, q.headers, idx))
	}
}
