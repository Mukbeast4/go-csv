package compress

import "strings"

type Format int

const (
	FormatNone Format = iota
	FormatGzip
	FormatBzip2
)

func (f Format) String() string {
	switch f {
	case FormatGzip:
		return "gzip"
	case FormatBzip2:
		return "bzip2"
	default:
		return "none"
	}
}

func DetectFormat(path string) Format {
	lower := strings.ToLower(path)
	switch {
	case strings.HasSuffix(lower, ".gz"), strings.HasSuffix(lower, ".gzip"):
		return FormatGzip
	case strings.HasSuffix(lower, ".bz2"), strings.HasSuffix(lower, ".bzip2"):
		return FormatBzip2
	default:
		return FormatNone
	}
}
