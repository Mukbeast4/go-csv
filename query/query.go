package query

import (
	gocsv "github.com/mukbeast4/go-csv"
)

type Query struct {
	headers []string
	rows    [][]string
	err     error
}

type Row struct {
	values  []string
	headers []string
	index   map[string]int
}

func newRow(values, headers []string, index map[string]int) Row {
	return Row{values: values, headers: headers, index: index}
}

func (r Row) Get(col string) string {
	if r.index == nil {
		return ""
	}
	i, ok := r.index[col]
	if !ok || i >= len(r.values) {
		return ""
	}
	return r.values[i]
}

func (r Row) At(i int) string {
	if i < 0 || i >= len(r.values) {
		return ""
	}
	return r.values[i]
}

func (r Row) Values() []string {
	out := make([]string, len(r.values))
	copy(out, r.values)
	return out
}

func (r Row) Headers() []string {
	return r.headers
}

func (r Row) Len() int {
	return len(r.values)
}

func From(f *gocsv.File) *Query {
	if f == nil {
		return &Query{}
	}
	headers := f.Headers()
	rows, err := f.GetRows()
	if err != nil {
		return &Query{err: err}
	}
	return &Query{headers: headers, rows: rows}
}

func FromRows(rows [][]string, headers []string) *Query {
	h := make([]string, len(headers))
	copy(h, headers)
	cp := make([][]string, len(rows))
	for i, r := range rows {
		row := make([]string, len(r))
		copy(row, r)
		cp[i] = row
	}
	return &Query{headers: h, rows: cp}
}

func FromIterator(it *gocsv.RowIterator) *Query {
	if it == nil {
		return &Query{}
	}
	headers := it.Headers()
	var rows [][]string
	for it.Next() {
		rows = append(rows, it.Row())
	}
	if err := it.Error(); err != nil {
		return &Query{err: err}
	}
	return &Query{headers: headers, rows: rows}
}

func (q *Query) Err() error {
	return q.err
}

func (q *Query) Headers() []string {
	out := make([]string, len(q.headers))
	copy(out, q.headers)
	return out
}

func (q *Query) headerIndex() map[string]int {
	idx := make(map[string]int, len(q.headers))
	for i, h := range q.headers {
		idx[h] = i
	}
	return idx
}

func (q *Query) Clone() *Query {
	if q.err != nil {
		return &Query{err: q.err}
	}
	h := make([]string, len(q.headers))
	copy(h, q.headers)
	rows := make([][]string, len(q.rows))
	for i, r := range q.rows {
		row := make([]string, len(r))
		copy(row, r)
		rows[i] = row
	}
	return &Query{headers: h, rows: rows}
}
