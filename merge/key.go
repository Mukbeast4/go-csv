package merge

import (
	"strings"

	gocsv "github.com/mukbeast4/go-csv"
)

type KeyFunc func(row []string, headers []string) string

func On(col string) KeyFunc {
	return func(row []string, headers []string) string {
		for i, h := range headers {
			if h == col && i < len(row) {
				return row[i]
			}
		}
		return ""
	}
}

func OnComposite(cols ...string) KeyFunc {
	return func(row []string, headers []string) string {
		parts := make([]string, len(cols))
		for i, c := range cols {
			for j, h := range headers {
				if h == c && j < len(row) {
					parts[i] = row[j]
					break
				}
			}
		}
		return strings.Join(parts, "\x1f")
	}
}

func buildIndex(f *gocsv.File, key KeyFunc) (map[string][][]string, []string, error) {
	headers := f.Headers()
	rows, err := f.GetRows()
	if err != nil {
		return nil, nil, err
	}
	idx := make(map[string][][]string)
	for _, row := range rows {
		k := key(row, headers)
		idx[k] = append(idx[k], row)
	}
	return idx, headers, nil
}
