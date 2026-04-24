package gocsv

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type CellType int

const (
	CellTypeEmpty CellType = iota
	CellTypeString
	CellTypeInt
	CellTypeFloat
	CellTypeBool
	CellTypeDate
)

func (ct CellType) String() string {
	switch ct {
	case CellTypeEmpty:
		return "empty"
	case CellTypeString:
		return "string"
	case CellTypeInt:
		return "int"
	case CellTypeFloat:
		return "float"
	case CellTypeBool:
		return "bool"
	case CellTypeDate:
		return "date"
	default:
		return "unknown"
	}
}

var dateFormats = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
	"02/01/2006",
	"01/02/2006",
	"2006/01/02",
}

func detectIntType(value any) (CellType, string, bool) {
	switch v := value.(type) {
	case int:
		return CellTypeInt, strconv.FormatInt(int64(v), 10), true
	case int8:
		return CellTypeInt, strconv.FormatInt(int64(v), 10), true
	case int16:
		return CellTypeInt, strconv.FormatInt(int64(v), 10), true
	case int32:
		return CellTypeInt, strconv.FormatInt(int64(v), 10), true
	case int64:
		return CellTypeInt, strconv.FormatInt(v, 10), true
	case uint:
		return CellTypeInt, strconv.FormatUint(uint64(v), 10), true
	case uint8:
		return CellTypeInt, strconv.FormatUint(uint64(v), 10), true
	case uint16:
		return CellTypeInt, strconv.FormatUint(uint64(v), 10), true
	case uint32:
		return CellTypeInt, strconv.FormatUint(uint64(v), 10), true
	case uint64:
		return CellTypeInt, strconv.FormatUint(v, 10), true
	default:
		return CellTypeEmpty, "", false
	}
}

func detectCellType(value any) (CellType, string) {
	switch v := value.(type) {
	case nil:
		return CellTypeEmpty, ""
	case string:
		if v == "" {
			return CellTypeEmpty, ""
		}
		return CellTypeString, v
	case float32:
		return CellTypeFloat, strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return CellTypeFloat, strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if v {
			return CellTypeBool, "true"
		}
		return CellTypeBool, "false"
	case time.Time:
		return CellTypeDate, v.Format(time.RFC3339)
	case []byte:
		return CellTypeString, string(v)
	case fmt.Stringer:
		return CellTypeString, v.String()
	default:
		if ct, s, ok := detectIntType(value); ok {
			return ct, s
		}
		return CellTypeString, fmt.Sprintf("%v", v)
	}
}

func inferCellType(raw string) CellType {
	if raw == "" {
		return CellTypeEmpty
	}
	if _, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return CellTypeInt
	}
	if _, err := strconv.ParseFloat(raw, 64); err == nil {
		return CellTypeFloat
	}
	lower := strings.ToLower(raw)
	if lower == "true" || lower == "false" {
		return CellTypeBool
	}
	for _, f := range dateFormats {
		if _, err := time.Parse(f, raw); err == nil {
			return CellTypeDate
		}
	}
	return CellTypeString
}

func parseInt(raw string) (int64, error) {
	if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return n, nil
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, err
	}
	return int64(f), nil
}

func parseFloat(raw string) (float64, error) {
	return strconv.ParseFloat(raw, 64)
}

func parseBool(raw string) (bool, error) {
	return strconv.ParseBool(raw)
}

func parseDate(raw string) (time.Time, error) {
	for _, f := range dateFormats {
		if t, err := time.Parse(f, raw); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("gocsv: cannot parse date %q", raw)
}

func valueToString(value any) (string, error) {
	_, s := detectCellType(value)
	if value == nil {
		return "", nil
	}
	return s, nil
}
