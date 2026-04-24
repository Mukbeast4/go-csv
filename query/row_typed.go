package query

import (
	"strconv"
	"strings"
	"time"
)

func (r Row) Int(col string) int64 {
	n, _ := strconv.ParseInt(strings.TrimSpace(r.Get(col)), 10, 64)
	return n
}

func (r Row) Float(col string) float64 {
	n, _ := strconv.ParseFloat(strings.TrimSpace(r.Get(col)), 64)
	return n
}

func (r Row) Bool(col string) bool {
	b, _ := strconv.ParseBool(strings.TrimSpace(r.Get(col)))
	return b
}

func (r Row) Date(col string) time.Time {
	raw := strings.TrimSpace(r.Get(col))
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"02/01/2006",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, raw); err == nil {
			return t
		}
	}
	return time.Time{}
}

func (r Row) Str(col string) string {
	return r.Get(col)
}
