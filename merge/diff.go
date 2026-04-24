package merge

import (
	gocsv "github.com/mukbeast4/go-csv"
)

type DiffResult struct {
	Added    [][]string
	Removed  [][]string
	Modified []ModifiedRow
	Headers  []string
}

type ModifiedRow struct {
	Key    string
	Before []string
	After  []string
	Fields []string
}

func Diff(before, after *gocsv.File, key KeyFunc) (*DiffResult, error) {
	headersBefore := before.Headers()
	headersAfter := after.Headers()
	rowsBefore, err := before.GetRows()
	if err != nil {
		return nil, err
	}
	rowsAfter, err := after.GetRows()
	if err != nil {
		return nil, err
	}

	beforeMap := make(map[string][]string)
	for _, r := range rowsBefore {
		beforeMap[key(r, headersBefore)] = r
	}
	afterMap := make(map[string][]string)
	for _, r := range rowsAfter {
		afterMap[key(r, headersAfter)] = r
	}

	result := &DiffResult{Headers: headersAfter}

	for k, rB := range beforeMap {
		rA, ok := afterMap[k]
		if !ok {
			result.Removed = append(result.Removed, rB)
			continue
		}
		changed := diffFields(rB, headersBefore, rA, headersAfter)
		if len(changed) > 0 {
			result.Modified = append(result.Modified, ModifiedRow{
				Key:    k,
				Before: rB,
				After:  rA,
				Fields: changed,
			})
		}
	}

	for k, rA := range afterMap {
		if _, ok := beforeMap[k]; !ok {
			result.Added = append(result.Added, rA)
		}
	}

	return result, nil
}

func diffFields(rB, hB, rA, hA []string) []string {
	var changed []string
	for i, h := range hB {
		var vB, vA string
		if i < len(rB) {
			vB = rB[i]
		}
		for j, ha := range hA {
			if ha == h && j < len(rA) {
				vA = rA[j]
				break
			}
		}
		if vB != vA {
			changed = append(changed, h)
		}
	}
	return changed
}
