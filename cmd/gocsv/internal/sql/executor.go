package sql

import (
	"fmt"
	"strconv"

	gocsv "github.com/mukbeast4/go-csv"
	"github.com/mukbeast4/go-csv/query"
)

func Execute(s *Statement, f *gocsv.File) (*gocsv.File, error) {
	q := query.From(f)
	if s.Where != nil {
		q = q.Where(func(r query.Row) bool {
			ok, _ := evalPredicate(s.Where, r)
			return ok
		})
	}
	if err := q.Err(); err != nil {
		return nil, err
	}

	if len(s.GroupBy) > 0 || hasAggregate(s.Select) {
		return executeAggregate(s, q)
	}

	for _, ord := range s.OrderBy {
		dir := query.Asc
		if ord.Desc {
			dir = query.Desc
		}
		q = q.OrderBy(ord.Column, dir)
	}
	if s.HasLimit {
		q = q.Limit(s.Limit)
	}

	if !isStarOnly(s.Select) {
		cols := make([]string, len(s.Select))
		aliases := make([]string, len(s.Select))
		for i, p := range s.Select {
			if p.Star {
				return nil, fmt.Errorf("cannot mix * with other columns")
			}
			if p.Agg != "" {
				return nil, fmt.Errorf("aggregate without GROUP BY requires all projections to be aggregates")
			}
			cols[i] = p.Column
			aliases[i] = p.Alias
		}
		q = q.Select(cols...)
		if err := q.Err(); err != nil {
			return nil, err
		}
		for i, alias := range aliases {
			if alias != "" {
				q = q.Rename(cols[i], alias)
			}
		}
	}

	if err := q.Err(); err != nil {
		return nil, err
	}
	return q.ToFile(), nil
}

func hasAggregate(projs []Projection) bool {
	for _, p := range projs {
		if p.Agg != "" {
			return true
		}
	}
	return false
}

func isStarOnly(projs []Projection) bool {
	return len(projs) == 1 && projs[0].Star
}

func executeAggregate(s *Statement, q *query.Query) (*gocsv.File, error) {
	headers := buildAggHeaders(s.Select)
	out := gocsv.NewFile()
	out.SetHeaders(headers)

	if len(s.GroupBy) == 0 {
		row := computeAggRow(s.Select, q, nil)
		out.AppendStrRow(row)
		return applyOrderAndLimit(out, s)
	}

	groups := groupRows(q, s.GroupBy)
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sortKeys(keys)

	for _, k := range keys {
		g := groups[k]
		row := computeAggRow(s.Select, g, s.GroupBy)
		out.AppendStrRow(row)
	}
	return applyOrderAndLimit(out, s)
}

func buildAggHeaders(projs []Projection) []string {
	out := make([]string, len(projs))
	for i, p := range projs {
		switch {
		case p.Alias != "":
			out[i] = p.Alias
		case p.Star:
			out[i] = "*"
		case p.Agg != "":
			if p.Column == "*" {
				out[i] = p.Agg + "(*)"
			} else {
				out[i] = p.Agg + "(" + p.Column + ")"
			}
		default:
			out[i] = p.Column
		}
	}
	return out
}

func computeAggRow(projs []Projection, q *query.Query, groupCols []string) []string {
	row := make([]string, len(projs))
	firstRow, hasFirst := q.First()
	for i, p := range projs {
		switch {
		case p.Star:
			if hasFirst {
				row[i] = joinRow(firstRow.Values())
			}
		case p.Agg != "":
			row[i] = applyAggregate(p.Agg, p.Column, q)
		default:
			if isGroupCol(p.Column, groupCols) && hasFirst {
				row[i] = firstRow.Get(p.Column)
			} else if hasFirst {
				row[i] = firstRow.Get(p.Column)
			}
		}
	}
	return row
}

func applyAggregate(agg, col string, q *query.Query) string {
	switch agg {
	case "COUNT":
		if col == "*" {
			return strconv.Itoa(q.Count())
		}
		return strconv.Itoa(countNonEmpty(q, col))
	case "SUM":
		return formatFloat(q.Sum(col))
	case "AVG":
		return formatFloat(q.Avg(col))
	case "MIN":
		return formatFloat(q.Min(col))
	case "MAX":
		return formatFloat(q.Max(col))
	}
	return ""
}

func countNonEmpty(q *query.Query, col string) int {
	count := 0
	q.ForEach(func(r query.Row) {
		if r.Get(col) != "" {
			count++
		}
	})
	return count
}

func formatFloat(v float64) string {
	if v == float64(int64(v)) {
		return strconv.FormatInt(int64(v), 10)
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func isGroupCol(col string, groupCols []string) bool {
	for _, g := range groupCols {
		if g == col {
			return true
		}
	}
	return false
}

func groupRows(q *query.Query, cols []string) map[string]*query.Query {
	if len(cols) == 1 {
		return q.GroupBy(cols[0])
	}
	rows := q.ToRows()
	headers := q.Headers()
	idx := make([]int, len(cols))
	for i, c := range cols {
		idx[i] = -1
		for j, h := range headers {
			if h == c {
				idx[i] = j
				break
			}
		}
	}
	groups := make(map[string][][]string)
	for _, r := range rows {
		parts := make([]string, len(idx))
		for i, ci := range idx {
			if ci >= 0 && ci < len(r) {
				parts[i] = r[ci]
			}
		}
		key := joinKey(parts)
		groups[key] = append(groups[key], r)
	}
	out := make(map[string]*query.Query, len(groups))
	for k, grows := range groups {
		out[k] = query.FromRows(grows, headers)
	}
	return out
}

func joinKey(parts []string) string {
	total := len(parts)
	for _, p := range parts {
		total += len(p)
	}
	b := make([]byte, 0, total)
	for i, p := range parts {
		if i > 0 {
			b = append(b, 0x1f)
		}
		b = append(b, p...)
	}
	return string(b)
}

func joinRow(fields []string) string {
	total := len(fields)
	for _, f := range fields {
		total += len(f)
	}
	b := make([]byte, 0, total)
	for i, f := range fields {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, f...)
	}
	return string(b)
}

func sortKeys(keys []string) {
	for i := 1; i < len(keys); i++ {
		k := keys[i]
		j := i
		for j > 0 && keys[j-1] > k {
			keys[j] = keys[j-1]
			j--
		}
		keys[j] = k
	}
}

func applyOrderAndLimit(f *gocsv.File, s *Statement) (*gocsv.File, error) {
	q := query.From(f)
	for _, ord := range s.OrderBy {
		dir := query.Asc
		if ord.Desc {
			dir = query.Desc
		}
		q = q.OrderBy(ord.Column, dir)
	}
	if s.HasLimit {
		q = q.Limit(s.Limit)
	}
	if err := q.Err(); err != nil {
		return nil, err
	}
	return q.ToFile(), nil
}
