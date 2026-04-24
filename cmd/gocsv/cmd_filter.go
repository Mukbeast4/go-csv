package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	gocsv "github.com/mukbeast4/go-csv"
	"github.com/mukbeast4/go-csv/query"
)

func cmdFilter(args []string) error {
	fs := flag.NewFlagSet("filter", flag.ContinueOnError)
	where := fs.String("w", "", "filter expression: col op value (ops: ==, !=, <, >, <=, >=, contains, starts, regex)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *where == "" || fs.NArg() < 1 {
		return fmt.Errorf("usage: filter -w \"col op value\" file.csv")
	}
	pred, err := parseExpr(*where)
	if err != nil {
		return err
	}
	f, err := gocsv.OpenFile(fs.Arg(0), gocsv.WithHeader(true))
	if err != nil {
		return err
	}
	out := query.From(f).Where(pred).ToFile()
	if out == nil {
		return fmt.Errorf("query error")
	}
	return out.Write(os.Stdout)
}

var opList = []string{"<=", ">=", "==", "!=", "<", ">", "contains", "starts", "regex"}

func parseExpr(expr string) (func(query.Row) bool, error) {
	expr = strings.TrimSpace(expr)
	for _, op := range opList {
		idx := findOp(expr, op)
		if idx == -1 {
			continue
		}
		col := strings.TrimSpace(expr[:idx])
		val := strings.TrimSpace(expr[idx+len(op):])
		val = strings.Trim(val, `"'`)
		return buildPred(col, op, val)
	}
	return nil, fmt.Errorf("invalid expression: %q", expr)
}

func findOp(expr, op string) int {
	if op == "contains" || op == "starts" || op == "regex" {
		idx := strings.Index(expr, " "+op+" ")
		if idx == -1 {
			return -1
		}
		return idx + 1
	}
	return strings.Index(expr, op)
}

func buildPred(col, op, val string) (func(query.Row) bool, error) {
	numVal, numErr := strconv.ParseFloat(val, 64)
	switch op {
	case "==":
		return func(r query.Row) bool { return r.Get(col) == val }, nil
	case "!=":
		return func(r query.Row) bool { return r.Get(col) != val }, nil
	case "<":
		if numErr == nil {
			return func(r query.Row) bool { return r.Float(col) < numVal }, nil
		}
		return func(r query.Row) bool { return r.Get(col) < val }, nil
	case ">":
		if numErr == nil {
			return func(r query.Row) bool { return r.Float(col) > numVal }, nil
		}
		return func(r query.Row) bool { return r.Get(col) > val }, nil
	case "<=":
		if numErr == nil {
			return func(r query.Row) bool { return r.Float(col) <= numVal }, nil
		}
		return func(r query.Row) bool { return r.Get(col) <= val }, nil
	case ">=":
		if numErr == nil {
			return func(r query.Row) bool { return r.Float(col) >= numVal }, nil
		}
		return func(r query.Row) bool { return r.Get(col) >= val }, nil
	case "contains":
		return func(r query.Row) bool { return strings.Contains(r.Get(col), val) }, nil
	case "starts":
		return func(r query.Row) bool { return strings.HasPrefix(r.Get(col), val) }, nil
	case "regex":
		re, err := regexp.Compile(val)
		if err != nil {
			return nil, err
		}
		return func(r query.Row) bool { return re.MatchString(r.Get(col)) }, nil
	}
	return nil, fmt.Errorf("unknown op: %s", op)
}
