package sql

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/mukbeast4/go-csv/query"
)

func evalPredicate(e Expr, r query.Row) (bool, error) {
	if e == nil {
		return true, nil
	}
	switch v := e.(type) {
	case *BinaryExpr:
		if v.Op == "AND" {
			l, err := evalPredicate(v.Left, r)
			if err != nil || !l {
				return l, err
			}
			return evalPredicate(v.Right, r)
		}
		if v.Op == "OR" {
			l, err := evalPredicate(v.Left, r)
			if err != nil || l {
				return l, err
			}
			return evalPredicate(v.Right, r)
		}
		return evalCompare(v, r)
	case *NotExpr:
		inner, err := evalPredicate(v.Inner, r)
		return !inner, err
	}
	return false, fmt.Errorf("unsupported expression")
}

func evalCompare(e *BinaryExpr, r query.Row) (bool, error) {
	left := evalValue(e.Left, r)
	right := evalValue(e.Right, r)
	switch e.Op {
	case "=", "==":
		return left == right, nil
	case "!=", "<>":
		return left != right, nil
	case "<":
		return cmp(left, right) < 0, nil
	case ">":
		return cmp(left, right) > 0, nil
	case "<=":
		return cmp(left, right) <= 0, nil
	case ">=":
		return cmp(left, right) >= 0, nil
	case "LIKE":
		return likeMatch(left, right), nil
	case "IS":
		return strings.TrimSpace(left) == "", nil
	case "IS NOT":
		return strings.TrimSpace(left) != "", nil
	}
	return false, fmt.Errorf("unknown operator %q", e.Op)
}

func evalValue(e Expr, r query.Row) string {
	switch v := e.(type) {
	case *ColumnRef:
		return r.Get(v.Name)
	case *Literal:
		return v.Value
	}
	return ""
}

func cmp(a, b string) int {
	af, aerr := strconv.ParseFloat(strings.TrimSpace(a), 64)
	bf, berr := strconv.ParseFloat(strings.TrimSpace(b), 64)
	if aerr == nil && berr == nil {
		switch {
		case af < bf:
			return -1
		case af > bf:
			return 1
		default:
			return 0
		}
	}
	return strings.Compare(a, b)
}

func likeMatch(s, pattern string) bool {
	re, err := likeToRegex(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(s)
}

func likeToRegex(pattern string) (*regexp.Regexp, error) {
	var b strings.Builder
	b.WriteString("^")
	for _, r := range pattern {
		switch r {
		case '%':
			b.WriteString(".*")
		case '_':
			b.WriteString(".")
		case '.', '+', '*', '?', '(', ')', '[', ']', '{', '}', '|', '^', '$', '\\':
			b.WriteRune('\\')
			b.WriteRune(r)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteString("$")
	return regexp.Compile(b.String())
}
