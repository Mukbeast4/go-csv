package query

import (
	"sort"
	"strconv"
	"strings"
)

type SortDir int

const (
	Asc SortDir = iota
	Desc
)

func (q *Query) OrderBy(col string, dir SortDir) *Query {
	if q.err != nil {
		return q
	}
	idx, ok := q.headerIndex()[col]
	if !ok {
		return &Query{err: &MissingColumnError{Column: col}}
	}
	rows := make([][]string, len(q.rows))
	copy(rows, q.rows)
	sort.SliceStable(rows, func(i, j int) bool {
		a := fieldAt(rows[i], idx)
		b := fieldAt(rows[j], idx)
		less := compareField(a, b)
		if dir == Desc {
			return less > 0
		}
		return less < 0
	})
	return &Query{headers: q.headers, rows: rows}
}

func (q *Query) OrderByFn(less func(a, b Row) bool) *Query {
	if q.err != nil {
		return q
	}
	idx := q.headerIndex()
	rows := make([][]string, len(q.rows))
	copy(rows, q.rows)
	sort.SliceStable(rows, func(i, j int) bool {
		return less(newRow(rows[i], q.headers, idx), newRow(rows[j], q.headers, idx))
	})
	return &Query{headers: q.headers, rows: rows}
}

func fieldAt(row []string, idx int) string {
	if idx < len(row) {
		return row[idx]
	}
	return ""
}

func compareField(a, b string) int {
	if af, err := strconv.ParseFloat(a, 64); err == nil {
		if bf, err := strconv.ParseFloat(b, 64); err == nil {
			switch {
			case af < bf:
				return -1
			case af > bf:
				return 1
			default:
				return 0
			}
		}
	}
	return strings.Compare(a, b)
}
