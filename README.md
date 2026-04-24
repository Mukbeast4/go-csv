# go-csv

[![CI](https://github.com/Mukbeast4/go-csv/actions/workflows/ci.yml/badge.svg)](https://github.com/Mukbeast4/go-csv/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/mukbeast4/go-csv.svg)](https://pkg.go.dev/github.com/mukbeast4/go-csv)
[![Go Report Card](https://goreportcard.com/badge/github.com/mukbeast4/go-csv)](https://goreportcard.com/report/github.com/mukbeast4/go-csv)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/Mukbeast4/go-csv)](https://github.com/Mukbeast4/go-csv/releases)

Pure Go, zero-dependency CSV library. Streaming reads and writes, typed cell accessors, header-based access, A1 notation, automatic delimiter and encoding detection.

## Features

- Read and write CSV, TSV, or any delimited format
- Streaming read (`RowIterator`, `StreamReader`) and streaming write (`StreamWriter`) for files of any size
- Typed cell accessors (`SetCellStr`, `GetCellInt`, `GetCellFloat`, `GetCellBool`, `GetCellDate`)
- Header-based record API (`GetByHeader`, `AppendRecord`, `GetRecords`)
- A1 notation and `(col, row)` coordinate access (`Cell`, `Range("A1:C10")`)
- Automatic delimiter and encoding detection
- UTF-8, UTF-16 LE/BE, ISO-8859-1, Windows-1252 with BOM handling
- RFC 4180 compliant, with lazy-quotes and error-recovery modes (`Strict`, `Skip`, `Collect`)
- `ParseError` with line, column, and byte offset for precise diagnostics
- Zero external dependencies (stdlib only, no `require` in `go.mod`)

## Requirements

- Go 1.25+

## Installation

```
go get github.com/mukbeast4/go-csv
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/mukbeast4/go-csv"
)

func main() {
    f, err := gocsv.OpenFile("data.csv", gocsv.WithHeader(true))
    if err != nil {
        panic(err)
    }

    fmt.Println(f.Headers())

    records, _ := f.GetRecords()
    for _, r := range records {
        fmt.Println(r["name"], r["age"])
    }

    f.AppendRecord(map[string]any{"name": "Charlie", "age": 28})
    f.SaveAs("data_out.csv")
}
```

## Reading

```go
f, _ := gocsv.OpenFile("data.csv", gocsv.WithHeader(true))

rows, _ := f.GetRows()
for i, row := range rows {
    fmt.Println(i, row)
}

v, _ := f.GetCellStr("B2")
n, _ := f.GetCellInt("C3")
x, _ := f.GetCellFloat("D4")

v, _ = f.GetByHeader(0, "email")
```

## Writing

```go
f := gocsv.NewFile()
f.SetHeaders([]string{"id", "name", "score"})
f.AppendRow([]any{1, "Alice", 95.5})
f.AppendRow([]any{2, "Bob", 87.3})
f.SaveAs("out.csv")
```

## Streaming Large Files

Read any size without loading the whole file:

```go
it, _ := gocsv.StreamReaderFromFile("huge.csv", gocsv.WithHeader(true))
defer it.Close()

for it.Next() {
    rec := it.Record()
    process(rec)
}
if err := it.Error(); err != nil {
    log.Fatal(err)
}
```

Write any size:

```go
sw, _ := gocsv.NewStreamWriterToFile("huge_out.csv")
defer sw.Close()

sw.WriteHeader([]string{"id", "value"})
for i := 0; i < 1_000_000; i++ {
    sw.WriteRow([]any{i, rand.Float64()})
}
```

## Configuration

| Option | Default | Description |
|---|---|---|
| `WithDelimiter(r)` | auto-detect | Field delimiter |
| `WithQuote(r)` | `"` | Quote character |
| `WithComment(r)` | disabled | Comment prefix (lines starting with this are skipped) |
| `WithHeader(b)` | auto-detect | Whether first row is a header |
| `WithEncoding(e)` | auto-detect | `EncodingUTF8`, `EncodingUTF16LE`, `EncodingUTF16BE`, `EncodingISO88591`, `EncodingWindows1252` |
| `WithLazyQuotes(b)` | `false` | Tolerate bare quotes in unquoted fields |
| `WithTrimLeadingSpace(b)` | `false` | Trim leading spaces in fields |
| `WithCRLF(b)` | `false` | Use `\r\n` line endings on write |
| `WithErrorMode(m)` | `Strict` | `Strict`, `Skip`, or `Collect` bad rows |
| `WithFieldsPerRecord(n)` | `0` (any) | Enforce column count (`-1` = strict first-row match) |
| `WithSkipRows(n)` | `0` | Skip N lines before parsing |
| `WithStdlibParser()` | off | Use `encoding/csv` as parser (fallback) |
| `WithWriteBOM(b)` | `false` | Write UTF-8 BOM on output |

## Error Handling

```go
f, err := gocsv.OpenBytes(data, gocsv.WithErrorMode(gocsv.ErrorModeCollect))
if err != nil {
    log.Fatal(err)
}

for _, pe := range f.ParseErrors() {
    fmt.Printf("line %d col %d offset %d: %v\n", pe.Line, pe.Column, pe.Offset, pe.Err)
}
```

## Encoding

UTF-8 with optional BOM is the default. Legacy files are read transparently when a BOM is present; otherwise pass `WithEncoding`:

```go
f, _ := gocsv.OpenFile("legacy.csv", gocsv.WithEncoding(gocsv.EncodingWindows1252))
```

## Range API

```go
f.Range("A1:C10").ForEach(func(col, row int, v string) {
    fmt.Printf("%d,%d => %s\n", col, row, v)
})

f.Range("A1:C10").SetValue("default")
f.Range("A1:B2").SetValues([][]any{{"a", "b"}, {"c", "d"}})
```

## Performance

Benchmarks on Apple M4 Pro (Apple Silicon, 12 cores):

```
10k rows (~580 KB):
  BenchmarkOpenBytes          1.3 ms/op    411 MB/s   (safe, default)
  BenchmarkOpenBytesUnsafe    0.9 ms/op    577 MB/s   (WithUnsafeStrings)
  BenchmarkOpenBytesStdlib    0.9 ms/op    584 MB/s   (encoding/csv)

200k rows (~11 MB, triggers parallel):
  BenchmarkOpenBytesParallel   10 ms/op   1163 MB/s   (WithUnsafeStrings, auto-parallel)
  BenchmarkOpenBytesSequential 17 ms/op    701 MB/s   (WithUnsafeStrings, serial)
```

### Fast options

- **Default** — safe, 71 % of stdlib throughput, half the allocations.
- **`WithUnsafeStrings()`** — uses `unsafe.String` for zero-copy fields, reaches ~97 % of stdlib throughput with half as many allocations. **Caller must not modify the input `[]byte` while the `*File` is in use** — fields share memory with the input buffer.
- **Parallel parsing** — automatically enabled when input `[]byte` is >= 10 MB. Uses `runtime.NumCPU()` workers by default. Doubles stdlib throughput on large files. Configure with `WithParallel(n)` to force N workers (`n=1` disables, `n=0` auto) and `WithParallelThreshold(bytes)` to tune the auto-activation size.
- **`WithStdlibParser()`** — delegate to `encoding/csv`. Loses BOM/encoding/error-mode features but matches stdlib exactly.

## Concurrency

`File` is not safe for concurrent mutation. Read-only access after the file is fully loaded is safe.

## v2 Subpackages

v2 adds opt-in subpackages for common workflows. The core `gocsv` package remains zero-dependency; subpackages are imported only when used.

### `query` — Filter / Map / Sort / Aggregate

Fluent, chainable API over `*gocsv.File`.

```go
import "github.com/mukbeast4/go-csv/query"

result := query.From(file).
    Where(func(r query.Row) bool { return r.Int("age") >= 30 }).
    Select("name", "email").
    OrderBy("name", query.Asc).
    Limit(100)

count := result.Count()
avg   := query.From(file).Avg("score")
byCity := query.From(file).CountBy("city")
groups := query.From(file).GroupBy("status")
```

### `marshal` — Struct ↔ CSV type-safe

```go
import "github.com/mukbeast4/go-csv/marshal"

type User struct {
    ID      int       `csv:"id"`
    Name    string    `csv:"name,required"`
    Email   string    `csv:"email,omitempty"`
    Created time.Time `csv:"created,format=2006-01-02"`
    Tags    []string  `csv:"tags,sep=|"`
}

var users []User
marshal.UnmarshalFile("users.csv", &users)
marshal.MarshalFile("out.csv", users)

enc, _ := marshal.NewEncoder(w, User{})
enc.EncodeAll(users)
dec, _ := marshal.NewDecoder(r, User{})
var u User
dec.Decode(&u)
```

Tag options: `name`, `required`, `omitempty`, `format=<layout>`, `sep=<rune>`, `-` to skip.

### `schema` — Declarative validation

```go
import "github.com/mukbeast4/go-csv/schema"

s := schema.New().
    Col("id", schema.Required(), schema.Int()).
    Col("email", schema.Regex(`^[^@]+@[^@]+$`)).
    Col("age", schema.Range(0, 150)).
    Col("role", schema.OneOf("admin", "user", "guest"))

for _, e := range s.Validate(file) {
    fmt.Printf("row %d field %q: %v\n", e.Row, e.Field, e.Err)
}
```

### `merge` — Join / Concat / Diff

```go
import "github.com/mukbeast4/go-csv/merge"

joined, _ := merge.LeftJoin(users, orders, merge.On("user_id"))
concat, _ := merge.Concat(jan, feb, mar)
union,  _ := merge.UnionBy(a, b, merge.On("id"))

d, _ := merge.Diff(before, after, merge.On("id"))
// d.Added / d.Removed / d.Modified
```

### `compress` — gzip / bzip2 transparent I/O

```go
import "github.com/mukbeast4/go-csv/compress"

f, _ := compress.Open("data.csv.gz")
compress.SaveAs(f, "out.csv.gz")

sw, closer, _ := compress.NewStreamWriter("huge.csv.gz")
defer closer.Close()
```

### `ods` — CSV ↔ ODS via go-ods

```go
import "github.com/mukbeast4/go-csv/ods"

ods.ToODS(csvFile, "report.ods", ods.WithSheetName("Data"))
f, _ := ods.FromODS("report.ods", "Sheet1")
ods.AppendSheet("report.ods", "Q2", csvFile)
```

### `cmd/gocsv` — CLI binary

```
go install github.com/mukbeast4/go-csv/cmd/gocsv@latest

gocsv head -n 20 file.csv
gocsv tail -n 10 file.csv
gocsv select -c name,email users.csv
gocsv filter -w "age > 30" users.csv
gocsv sort -c age -desc users.csv
gocsv stats file.csv
gocsv validate -schema schema.json file.csv
gocsv convert -to ods file.csv out.ods
gocsv join -on user_id -mode left a.csv b.csv
gocsv diff -on id before.csv after.csv
gocsv sql "SELECT city, COUNT(*), AVG(age) FROM t GROUP BY city" users.csv
gocsv gen-struct -name User -package models users.csv > user.go
```

**Filter ops**: `==`, `!=`, `<`, `>`, `<=`, `>=`, `contains`, `starts`, `regex`.

**SQL** supports: `SELECT cols | *`, aggregates (`COUNT/SUM/AVG/MIN/MAX`), `WHERE` (`=`, `!=`, `<`, `>`, `<=`, `>=`, `LIKE`, `IS NULL`, `IS NOT NULL`, `AND`, `OR`, `NOT`, parentheses), `GROUP BY`, `ORDER BY ... [ASC|DESC]`, `LIMIT n`, column aliases (`AS`). No JOIN or subqueries.

**gen-struct** scans the CSV, infers `int64 / float64 / bool / time.Time / string` per column, and emits a Go struct with `csv:` tags ready for the `marshal` package.

## Examples

Runnable examples in `examples/`:
```
go run ./examples/query
go run ./examples/marshal
go run ./examples/schema
go run ./examples/merge
go run ./examples/compress
```

## License

MIT
