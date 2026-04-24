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
- Zero external dependencies in the core package (stdlib only). Subpackages opt-in: `xlsx` uses excelize, `ods` uses go-ods, `compress` uses klauspost/compress for zstd.

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

Tag options: `name`, `required`, `omitempty`, `format=<layout>`, `sep=<rune>`, `flatten`, `prefix=<str>`, `json`, `rest`, `-` to skip.

**Custom types via interfaces** — implement `marshal.Marshaler` / `marshal.Unmarshaler` for CSV-specific serialization. Types implementing `encoding.TextMarshaler` / `encoding.TextUnmarshaler` work transparently (e.g. `*big.Int`, `net.IP`). Priority: CSV interface > Text interface > built-in.

```go
type Color struct{ R, G, B int }
func (c Color) MarshalCSV() (string, error)  { return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B), nil }
func (c *Color) UnmarshalCSV(s string) error { _, err := fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B); return err }
```

**Nested structs** — opt-in via `flatten` (multi-column) or `json` (single JSON-encoded cell):

```go
type Address struct {
    Street string `csv:"street"`
    City   string `csv:"city"`
}
type Employee struct {
    ID      int     `csv:"id"`
    Addr    Address `csv:"addr,flatten"`          // → addr.street, addr.city
    Home    Address `csv:"home,flatten,prefix=h_"` // → h_street, h_city
    Profile Profile `csv:"profile,json"`           // single JSON cell
}
```

**Rest map** — capture unmatched columns into a `map[string]string` via `csv:",rest"`:

```go
type Record struct {
    ID    int               `csv:"id"`
    Name  string            `csv:"name"`
    Extra map[string]string `csv:",rest"`
}
// On decode: unknown headers go into Extra.
// On encode (slice): union of all keys, sorted, appended as extra columns.
// Streaming Encoder: call SetRestKeys(...) before Encode.
```

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

### `compress` — gzip / bzip2 / zstd transparent I/O

```go
import "github.com/mukbeast4/go-csv/compress"

f, _ := compress.Open("data.csv.gz")   // .gz, .bz2 (read), .zst/.zstd
compress.SaveAs(f, "out.csv.zst")      // gz and zstd also supported for write

sw, closer, _ := compress.NewStreamWriter("huge.csv.zst")
defer closer.Close()
```

Format is auto-detected from the extension. zstd read/write uses `github.com/klauspost/compress/zstd` (pure Go, no cgo). bzip2 is read-only (stdlib limitation).

### `ods` — CSV ↔ ODS via go-ods

```go
import "github.com/mukbeast4/go-csv/ods"

ods.ToODS(csvFile, "report.ods", ods.WithSheetName("Data"))
f, _ := ods.FromODS("report.ods", "Sheet1")
ods.AppendSheet("report.ods", "Q2", csvFile)
```

### `xlsx` — CSV ↔ XLSX via excelize

```go
import "github.com/mukbeast4/go-csv/xlsx"

xlsx.ToXLSX(csvFile, "report.xlsx", xlsx.WithSheetName("Data"), xlsx.WithAutoFilter(true))
f, _ := xlsx.FromXLSX("report.xlsx", "Data")
xlsx.AppendSheet("report.xlsx", "Q2", csvFile)

names, _ := xlsx.SheetNames("report.xlsx")

// Streaming for large files (> ~100k rows)
sw, closer, _ := xlsx.NewStreamWriter("huge.xlsx", "Data")
defer closer.Close()
sw.WriteHeader([]string{"id", "value"})
for i := 0; i < 1_000_000; i++ {
    sw.WriteStrRow([]string{strconv.Itoa(i), "x"})
}

it, rcloser, _ := xlsx.NewStreamReader("huge.xlsx", "Data")
defer rcloser.Close()
for it.Next() {
    row := it.Row()
    _ = row
}
```

Depends on `github.com/xuri/excelize/v2`. Cell types are auto-inferred on write (ints, floats, bools, dates become typed Excel cells) — disable with `WithTypeInfer(false)`.

### `jsonx` — CSV ↔ JSON / NDJSON

```go
import "github.com/mukbeast4/go-csv/jsonx"

jsonx.ToJSON(f, os.Stdout, jsonx.WithPretty(true))   // [{...},{...}]
jsonx.ToNDJSON(f, os.Stdout)                          // {...}\n{...}

f, _ := jsonx.FromJSON(r)        // accepts [{...},{...}]
f, _ := jsonx.FromNDJSON(r)      // streams line-delimited objects

// Streaming NDJSON
enc := jsonx.NewEncoder(w, []string{"id", "name"})
enc.Encode([]string{"1", "Alice"})

it, _ := jsonx.StreamReader(r)  // returns a *gocsv.RowIterator
defer it.Close()
for it.Next() { ... }
```

Type inference is on by default (`"123"` → `123`, `"true"` → `true`). Header order is preserved in JSON output via a custom ordered-map marshaler. Large integers (>2^53) are preserved via `json.Decoder.UseNumber()` on decode.

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
gocsv convert -to xlsx file.csv out.xlsx
gocsv convert -to json file.csv out.json
gocsv convert -to ndjson file.csv out.ndjson
gocsv convert -to zst file.csv out.csv.zst
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
go run ./examples/jsonx
go run ./examples/xlsx
```

## License

MIT
