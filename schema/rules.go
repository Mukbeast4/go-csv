package schema

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Rule interface {
	Name() string
	Check(value string) error
}

type requiredRule struct{}

func (r *requiredRule) Name() string { return "required" }
func (r *requiredRule) Check(v string) error {
	if strings.TrimSpace(v) == "" {
		return fmt.Errorf("value is required")
	}
	return nil
}

func Required() Rule { return &requiredRule{} }

type typeRule struct {
	kind string
	fn   func(string) error
}

func (r *typeRule) Name() string { return "type:" + r.kind }
func (r *typeRule) Check(v string) error {
	if v == "" {
		return nil
	}
	return r.fn(v)
}

func Int() Rule {
	return &typeRule{kind: "int", fn: func(v string) error {
		_, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf("not an integer: %q", v)
		}
		return nil
	}}
}

func Float() Rule {
	return &typeRule{kind: "float", fn: func(v string) error {
		_, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return fmt.Errorf("not a float: %q", v)
		}
		return nil
	}}
}

func Bool() Rule {
	return &typeRule{kind: "bool", fn: func(v string) error {
		_, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("not a bool: %q", v)
		}
		return nil
	}}
}

func Date(layout string) Rule {
	return &typeRule{kind: "date", fn: func(v string) error {
		_, err := time.Parse(layout, v)
		if err != nil {
			return fmt.Errorf("not a date %s: %q", layout, v)
		}
		return nil
	}}
}

func String() Rule {
	return &typeRule{kind: "string", fn: func(v string) error { return nil }}
}

type regexRule struct {
	pattern *regexp.Regexp
	raw     string
}

func (r *regexRule) Name() string { return "regex" }
func (r *regexRule) Check(v string) error {
	if v == "" {
		return nil
	}
	if !r.pattern.MatchString(v) {
		return fmt.Errorf("does not match %s", r.raw)
	}
	return nil
}

func Regex(pattern string) Rule {
	return &regexRule{pattern: regexp.MustCompile(pattern), raw: pattern}
}

type rangeRule struct {
	min, max float64
	hasMin   bool
	hasMax   bool
}

func (r *rangeRule) Name() string { return "range" }
func (r *rangeRule) Check(v string) error {
	if v == "" {
		return nil
	}
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fmt.Errorf("not numeric: %q", v)
	}
	if r.hasMin && n < r.min {
		return fmt.Errorf("%v < min %v", n, r.min)
	}
	if r.hasMax && n > r.max {
		return fmt.Errorf("%v > max %v", n, r.max)
	}
	return nil
}

func Min(n float64) Rule {
	return &rangeRule{min: n, hasMin: true}
}

func Max(n float64) Rule {
	return &rangeRule{max: n, hasMax: true}
}

func Range(min, max float64) Rule {
	return &rangeRule{min: min, max: max, hasMin: true, hasMax: true}
}

type lengthRule struct {
	min, max int
	hasMin   bool
	hasMax   bool
}

func (r *lengthRule) Name() string { return "length" }
func (r *lengthRule) Check(v string) error {
	l := len([]rune(v))
	if r.hasMin && l < r.min {
		return fmt.Errorf("length %d < %d", l, r.min)
	}
	if r.hasMax && l > r.max {
		return fmt.Errorf("length %d > %d", l, r.max)
	}
	return nil
}

func MinLen(n int) Rule {
	return &lengthRule{min: n, hasMin: true}
}

func MaxLen(n int) Rule {
	return &lengthRule{max: n, hasMax: true}
}

type oneOfRule struct {
	values map[string]struct{}
	list   []string
}

func (r *oneOfRule) Name() string { return "oneof" }
func (r *oneOfRule) Check(v string) error {
	if v == "" {
		return nil
	}
	if _, ok := r.values[v]; !ok {
		return fmt.Errorf("must be one of %v", r.list)
	}
	return nil
}

func OneOf(values ...string) Rule {
	m := make(map[string]struct{}, len(values))
	for _, v := range values {
		m[v] = struct{}{}
	}
	return &oneOfRule{values: m, list: values}
}

type customRule struct {
	name string
	fn   func(string) error
}

func (r *customRule) Name() string         { return r.name }
func (r *customRule) Check(v string) error { return r.fn(v) }

func Custom(name string, fn func(string) error) Rule {
	return &customRule{name: name, fn: fn}
}
