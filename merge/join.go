package merge

import (
	gocsv "github.com/mukbeast4/go-csv"
)

func InnerJoin(a, b *gocsv.File, key KeyFunc) (*gocsv.File, error) {
	return doJoin(a, b, key, joinInner)
}

func LeftJoin(a, b *gocsv.File, key KeyFunc) (*gocsv.File, error) {
	return doJoin(a, b, key, joinLeft)
}

func RightJoin(a, b *gocsv.File, key KeyFunc) (*gocsv.File, error) {
	return doJoin(a, b, key, joinRight)
}

func FullJoin(a, b *gocsv.File, key KeyFunc) (*gocsv.File, error) {
	return doJoin(a, b, key, joinFull)
}

type joinMode int

const (
	joinInner joinMode = iota
	joinLeft
	joinRight
	joinFull
)

func doJoin(a, b *gocsv.File, key KeyFunc, mode joinMode) (*gocsv.File, error) {
	headersA := a.Headers()
	headersB := b.Headers()
	rowsA, err := a.GetRows()
	if err != nil {
		return nil, err
	}
	rowsB, err := b.GetRows()
	if err != nil {
		return nil, err
	}

	indexB := make(map[string][][]string)
	for _, r := range rowsB {
		k := key(r, headersB)
		indexB[k] = append(indexB[k], r)
	}

	usedB := make(map[string]bool)

	mergedHeaders := append([]string{}, headersA...)
	for _, h := range headersB {
		if !contains(headersA, h) {
			mergedHeaders = append(mergedHeaders, h)
		}
	}

	out := gocsv.NewFile()
	out.SetHeaders(mergedHeaders)

	emptyB := make([]string, len(headersB))
	emptyA := make([]string, len(headersA))

	for _, rA := range rowsA {
		k := key(rA, headersA)
		matches, ok := indexB[k]
		if ok {
			usedB[k] = true
			for _, rB := range matches {
				merged := mergeRow(rA, rB, headersA, headersB, mergedHeaders)
				out.AppendStrRow(merged)
			}
		} else if mode == joinLeft || mode == joinFull {
			merged := mergeRow(rA, emptyB, headersA, headersB, mergedHeaders)
			out.AppendStrRow(merged)
		}
	}

	if mode == joinRight || mode == joinFull {
		for _, rB := range rowsB {
			k := key(rB, headersB)
			if usedB[k] {
				continue
			}
			merged := mergeRow(emptyA, rB, headersA, headersB, mergedHeaders)
			out.AppendStrRow(merged)
		}
	}

	return out, nil
}

func mergeRow(rA, rB, headersA, headersB, merged []string) []string {
	out := make([]string, len(merged))
	for i, h := range merged {
		for j, hA := range headersA {
			if h == hA && j < len(rA) {
				out[i] = rA[j]
				break
			}
		}
		if out[i] != "" {
			continue
		}
		for j, hB := range headersB {
			if h == hB && j < len(rB) {
				out[i] = rB[j]
				break
			}
		}
	}
	return out
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
