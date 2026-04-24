package query

import "strconv"

func (q *Query) Count() int {
	if q.err != nil {
		return 0
	}
	return len(q.rows)
}

func (q *Query) Sum(col string) float64 {
	if q.err != nil {
		return 0
	}
	idx, ok := q.headerIndex()[col]
	if !ok {
		return 0
	}
	var sum float64
	for _, r := range q.rows {
		if idx >= len(r) {
			continue
		}
		if v, err := strconv.ParseFloat(r[idx], 64); err == nil {
			sum += v
		}
	}
	return sum
}

func (q *Query) Avg(col string) float64 {
	if q.err != nil || len(q.rows) == 0 {
		return 0
	}
	idx, ok := q.headerIndex()[col]
	if !ok {
		return 0
	}
	var sum float64
	var count int
	for _, r := range q.rows {
		if idx >= len(r) {
			continue
		}
		if v, err := strconv.ParseFloat(r[idx], 64); err == nil {
			sum += v
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

func (q *Query) Min(col string) float64 {
	if q.err != nil {
		return 0
	}
	idx, ok := q.headerIndex()[col]
	if !ok {
		return 0
	}
	var min float64
	first := true
	for _, r := range q.rows {
		if idx >= len(r) {
			continue
		}
		v, err := strconv.ParseFloat(r[idx], 64)
		if err != nil {
			continue
		}
		if first || v < min {
			min = v
			first = false
		}
	}
	return min
}

func (q *Query) Max(col string) float64 {
	if q.err != nil {
		return 0
	}
	idx, ok := q.headerIndex()[col]
	if !ok {
		return 0
	}
	var max float64
	first := true
	for _, r := range q.rows {
		if idx >= len(r) {
			continue
		}
		v, err := strconv.ParseFloat(r[idx], 64)
		if err != nil {
			continue
		}
		if first || v > max {
			max = v
			first = false
		}
	}
	return max
}

func (q *Query) CountBy(col string) map[string]int {
	out := make(map[string]int)
	if q.err != nil {
		return out
	}
	idx, ok := q.headerIndex()[col]
	if !ok {
		return out
	}
	for _, r := range q.rows {
		if idx >= len(r) {
			continue
		}
		out[r[idx]]++
	}
	return out
}

func (q *Query) GroupBy(col string) map[string]*Query {
	out := make(map[string]*Query)
	if q.err != nil {
		return out
	}
	idx, ok := q.headerIndex()[col]
	if !ok {
		return out
	}
	groups := make(map[string][][]string)
	for _, r := range q.rows {
		if idx >= len(r) {
			continue
		}
		groups[r[idx]] = append(groups[r[idx]], r)
	}
	for k, rows := range groups {
		headers := make([]string, len(q.headers))
		copy(headers, q.headers)
		out[k] = &Query{headers: headers, rows: rows}
	}
	return out
}

func (q *Query) Reduce(initial any, fn func(acc any, r Row) any) any {
	if q.err != nil {
		return initial
	}
	idx := q.headerIndex()
	acc := initial
	for _, r := range q.rows {
		acc = fn(acc, newRow(r, q.headers, idx))
	}
	return acc
}
