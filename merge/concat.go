package merge

import (
	gocsv "github.com/mukbeast4/go-csv"
)

func Concat(files ...*gocsv.File) (*gocsv.File, error) {
	if len(files) == 0 {
		return gocsv.NewFile(), nil
	}
	first := files[0]
	out := gocsv.NewFile()
	headers := first.Headers()
	if len(headers) > 0 {
		out.SetHeaders(headers)
	}
	for _, f := range files {
		fheaders := f.Headers()
		rows, err := f.GetRows()
		if err != nil {
			return nil, err
		}
		for _, r := range rows {
			if len(headers) == 0 {
				out.AppendStrRow(r)
				continue
			}
			aligned := alignRow(r, fheaders, headers)
			out.AppendStrRow(aligned)
		}
	}
	return out, nil
}

func alignRow(row, fromHeaders, toHeaders []string) []string {
	out := make([]string, len(toHeaders))
	for i, h := range toHeaders {
		for j, fh := range fromHeaders {
			if fh == h && j < len(row) {
				out[i] = row[j]
				break
			}
		}
	}
	return out
}

func UnionBy(a, b *gocsv.File, key KeyFunc) (*gocsv.File, error) {
	concat, err := Concat(a, b)
	if err != nil {
		return nil, err
	}
	headers := concat.Headers()
	rows, err := concat.GetRows()
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	out := gocsv.NewFile()
	out.SetHeaders(headers)
	for _, r := range rows {
		k := key(r, headers)
		if seen[k] {
			continue
		}
		seen[k] = true
		out.AppendStrRow(r)
	}
	return out, nil
}
