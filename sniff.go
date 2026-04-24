package gocsv

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/mukbeast4/go-csv/internal/encoding"
)

var candidateDelimiters = []rune{',', ';', '\t', '|'}

func SniffDelimiter(data []byte) rune {
	return sniffDelimiter(data)
}

func sniffDelimiter(data []byte) rune {
	if len(data) == 0 {
		return ','
	}
	sample := data
	if len(sample) > 8192 {
		sample = sample[:8192]
	}
	lines := splitSampleLines(sample, 10)
	if len(lines) == 0 {
		return ','
	}

	bestDelim := ','
	bestScore := -1

	for _, d := range candidateDelimiters {
		counts := make([]int, 0, len(lines))
		for _, line := range lines {
			counts = append(counts, countUnquoted(line, d))
		}
		first := counts[0]
		if first == 0 {
			continue
		}
		consistent := 0
		for _, c := range counts {
			if c == first {
				consistent++
			}
		}
		score := consistent*100 + first
		if score > bestScore {
			bestScore = score
			bestDelim = d
		}
	}
	return bestDelim
}

func splitSampleLines(data []byte, max int) []string {
	lines := bytes.Split(data, []byte{'\n'})
	result := make([]string, 0, len(lines))
	for _, l := range lines {
		trimmed := strings.TrimRight(string(l), "\r")
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
		if len(result) >= max {
			break
		}
	}
	return result
}

func countUnquoted(line string, delim rune) int {
	inQuote := false
	count := 0
	for _, r := range line {
		if r == '"' {
			inQuote = !inQuote
			continue
		}
		if r == delim && !inQuote {
			count++
		}
	}
	return count
}

func SniffEncoding(data []byte) Encoding {
	enc, _ := encoding.DetectBOM(data)
	return enc
}

func SniffHeader(rows [][]string) bool {
	return sniffHeader(rows)
}

func sniffHeader(rows [][]string) bool {
	if len(rows) < 2 {
		return false
	}
	first := rows[0]
	if len(first) == 0 {
		return false
	}
	for _, field := range first {
		if field == "" {
			return false
		}
		if _, err := strconv.ParseFloat(field, 64); err == nil {
			return false
		}
	}
	second := rows[1]
	numericInSecond := 0
	for _, field := range second {
		if _, err := strconv.ParseFloat(field, 64); err == nil {
			numericInSecond++
		}
	}
	if numericInSecond > 0 {
		return true
	}
	return false
}
