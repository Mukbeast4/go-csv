package parser

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/mukbeast4/go-csv/internal/dialect"
)

func genCSV(rows int) []byte {
	var buf bytes.Buffer
	buf.WriteString("id,name,score,city,note\n")
	for i := 0; i < rows; i++ {
		fmt.Fprintf(&buf, "%d,user%d,%.2f,city%d,\"note with, comma %d\"\n", i, i, float64(i)*1.5, i%100, i)
	}
	return buf.Bytes()
}

func TestParallelMatchesSequential(t *testing.T) {
	data := genCSV(5000)
	d := dialect.Default()
	seq, _, err := ParseBytes(data, d, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, workers := range []int{2, 4, 8} {
		par, _, err := ParseBytesParallel(data, d, false, workers)
		if err != nil {
			t.Fatalf("workers=%d: %v", workers, err)
		}
		if !reflect.DeepEqual(seq, par) {
			t.Fatalf("workers=%d: parallel result differs from sequential (seq=%d rows, par=%d rows)",
				workers, len(seq), len(par))
		}
	}
}

func TestParallelSmallFile(t *testing.T) {
	data := []byte("a,b\n1,2\n3,4\n")
	d := dialect.Default()
	par, _, err := ParseBytesParallel(data, d, false, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(par) != 3 {
		t.Errorf("rows: %d", len(par))
	}
}

func TestParallelQuotedNewlines(t *testing.T) {
	data := []byte("a,b\n1,\"embedded\nnewline\"\n3,normal\n")
	for range 50 {
		data = append(data, "x,\"more\nembedded\"\n"...)
	}
	d := dialect.Default()
	seq, _, _ := ParseBytes(data, d, false)
	par, _, err := ParseBytesParallel(data, d, false, 4)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(seq, par) {
		t.Errorf("parallel differs from sequential")
	}
}

func TestFindSafeNewline(t *testing.T) {
	data := []byte("a,b,c\n1,2,3\n4,5,6\n")
	n := findSafeNewline(data, 0, 6, '"')
	if n != 11 {
		t.Errorf("expected newline at 11, got %d", n)
	}
}

func TestFindSafeNewlineInsideQuote(t *testing.T) {
	data := []byte(`a,"quoted` + "\n" + `inside",c` + "\n" + `1,2,3` + "\n")
	n := findSafeNewline(data, 0, 10, '"')
	if n < 0 || data[n] != '\n' {
		t.Fatalf("not a newline at %d", n)
	}
	if n != 19 {
		t.Errorf("expected unquoted newline at 19, got %d", n)
	}
}

func TestSplitChunks(t *testing.T) {
	data := genCSV(1000)
	chunks := splitChunks(data, 4, '"')
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}
	total := 0
	for _, c := range chunks {
		total += len(c)
	}
	if total != len(data) {
		t.Errorf("chunk total %d != data %d", total, len(data))
	}
}
