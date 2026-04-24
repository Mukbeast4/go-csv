package parser

import (
	"errors"
	"runtime"
	"sync"

	"github.com/mukbeast4/go-csv/internal/dialect"
)

const MinParallelSize = 1 * 1024 * 1024

func ParseBytesParallel(data []byte, d dialect.Dialect, unsafeStrings bool, workers int) ([][]string, []*ParseError, error) {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if workers <= 1 || len(data) < MinParallelSize {
		return ParseBytes(data, d, unsafeStrings)
	}

	quote := byte(d.Quote)
	chunks := splitChunks(data, workers, quote)
	if len(chunks) <= 1 {
		return ParseBytes(data, d, unsafeStrings)
	}

	results := make([][][]string, len(chunks))
	errSlices := make([][]*ParseError, len(chunks))
	firstErrs := make([]*ParseError, len(chunks))

	var wg sync.WaitGroup
	for i, c := range chunks {
		wg.Add(1)
		go func(idx int, chunk []byte) {
			defer wg.Done()
			rows, errs, err := ParseBytes(chunk, d, unsafeStrings)
			results[idx] = rows
			errSlices[idx] = errs
			if err != nil {
				var pe *ParseError
				if errors.As(err, &pe) {
					firstErrs[idx] = pe
				} else {
					firstErrs[idx] = &ParseError{Err: err}
				}
			}
		}(i, c)
	}
	wg.Wait()

	totalRows := 0
	totalErrs := 0
	for i := range results {
		totalRows += len(results[i])
		totalErrs += len(errSlices[i])
	}
	out := make([][]string, 0, totalRows)
	var allErrs []*ParseError
	if totalErrs > 0 {
		allErrs = make([]*ParseError, 0, totalErrs)
	}
	for i := range results {
		out = append(out, results[i]...)
		allErrs = append(allErrs, errSlices[i]...)
	}

	for _, pe := range firstErrs {
		if pe != nil {
			if d.ErrorMode == dialect.ErrorModeStrict {
				return out, allErrs, pe
			}
		}
	}

	if d.FieldsPerRecord == -1 && len(out) > 1 {
		expected := len(out[0])
		for i := 1; i < len(out); i++ {
			if len(out[i]) != expected {
				pe := &ParseError{Line: i + 1, Err: ErrFieldCount}
				if d.ErrorMode == dialect.ErrorModeStrict {
					return out, allErrs, pe
				}
				allErrs = append(allErrs, pe)
			}
		}
	}

	return out, allErrs, nil
}

func splitChunks(data []byte, n int, quote byte) [][]byte {
	if n <= 1 || len(data) == 0 {
		return [][]byte{data}
	}
	approx := len(data) / n
	chunks := make([][]byte, 0, n)
	start := 0
	for i := 0; i < n-1; i++ {
		target := start + approx
		if target >= len(data) {
			break
		}
		cut := findSafeNewline(data, start, target, quote)
		if cut < 0 || cut <= start {
			break
		}
		chunks = append(chunks, data[start:cut+1])
		start = cut + 1
	}
	if start < len(data) {
		chunks = append(chunks, data[start:])
	}
	return chunks
}

func findSafeNewline(data []byte, chunkStart, target int, quote byte) int {
	inQuote := false
	for i := chunkStart; i < target && i < len(data); i++ {
		if data[i] == quote {
			inQuote = !inQuote
		}
	}
	for i := target; i < len(data); i++ {
		if data[i] == quote {
			inQuote = !inQuote
			continue
		}
		if data[i] == '\n' && !inQuote {
			return i
		}
	}
	return -1
}
