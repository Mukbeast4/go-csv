package gocsv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"testing"
)

func makeLargeCSV(rows int) []byte {
	var buf bytes.Buffer
	buf.WriteString("id,name,score,city,note\n")
	for i := 0; i < rows; i++ {
		fmt.Fprintf(&buf, "%d,user%d,%.2f,city%d,\"note with, comma %d\"\n", i, i, float64(i)*1.5, i%100, i)
	}
	return buf.Bytes()
}

func BenchmarkOpenBytes(b *testing.B) {
	data := makeLargeCSV(10000)
	b.ResetTimer()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		_, err := OpenBytes(data, WithHeader(true))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOpenBytesUnsafe(b *testing.B) {
	data := makeLargeCSV(10000)
	b.ResetTimer()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		_, err := OpenBytes(data, WithHeader(true), WithUnsafeStrings())
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOpenBytesParallel(b *testing.B) {
	data := makeLargeCSV(200000)
	b.ResetTimer()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		_, err := OpenBytes(data,
			WithHeader(true),
			WithUnsafeStrings(),
			WithParallelThreshold(1),
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOpenBytesSequentialLarge(b *testing.B) {
	data := makeLargeCSV(200000)
	b.ResetTimer()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		_, err := OpenBytes(data,
			WithHeader(true),
			WithUnsafeStrings(),
			WithParallel(1),
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkOpenBytesStdlib(b *testing.B) {
	data := makeLargeCSV(10000)
	b.ResetTimer()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		r := csv.NewReader(bytes.NewReader(data))
		for {
			_, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkStreamRead(b *testing.B) {
	data := makeLargeCSV(10000)
	b.ResetTimer()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		it, err := StreamReader(bytes.NewReader(data), WithHeader(true))
		if err != nil {
			b.Fatal(err)
		}
		for it.Next() {
		}
		it.Close()
	}
}

func BenchmarkStreamWrite(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		sw := NewStreamWriter(&buf)
		sw.WriteHeader([]string{"id", "name", "score"})
		for j := 0; j < 10000; j++ {
			sw.WriteRow([]any{j, fmt.Sprintf("user%d", j), float64(j) * 1.5})
		}
		sw.Close()
	}
}

func BenchmarkCellLookup(b *testing.B) {
	f, _ := OpenBytes(makeLargeCSV(1000), WithHeader(true))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = f.GetCellStr("B500")
	}
}

func BenchmarkSniffDelimiter(b *testing.B) {
	data := []byte(strings.Repeat("a,b,c\n1,2,3\n", 500))
	b.ResetTimer()
	b.SetBytes(int64(len(data)))
	for i := 0; i < b.N; i++ {
		SniffDelimiter(data)
	}
}
