package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIHead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")
	os.WriteFile(path, []byte("a,b\n1,2\n3,4\n5,6\n"), 0644)

	bin := buildCLI(t)
	out, err := exec.Command(bin, "head", "-n", "2", path).Output()
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) != 3 {
		t.Errorf("lines: %v", lines)
	}
}

func TestCLISelect(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")
	os.WriteFile(path, []byte("a,b,c\n1,2,3\n"), 0644)

	bin := buildCLI(t)
	out, err := exec.Command(bin, "select", "-c", "a,c", path).Output()
	if err != nil {
		t.Fatal(err)
	}
	s := strings.TrimSpace(string(out))
	if !strings.Contains(s, "a,c") {
		t.Errorf("header: %q", s)
	}
	if !strings.Contains(s, "1,3") {
		t.Errorf("data: %q", s)
	}
}

func TestCLIFilter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")
	os.WriteFile(path, []byte("id,age\n1,20\n2,30\n3,40\n"), 0644)

	bin := buildCLI(t)
	out, err := exec.Command(bin, "filter", "-w", "age > 25", path).Output()
	if err != nil {
		t.Fatal(err)
	}
	s := strings.TrimSpace(string(out))
	if strings.Contains(s, "20") {
		t.Errorf("should filter out 20: %q", s)
	}
	if !strings.Contains(s, "30") || !strings.Contains(s, "40") {
		t.Errorf("missing 30/40: %q", s)
	}
}

func TestCLIStats(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")
	os.WriteFile(path, []byte("id,name\n1,Alice\n2,Bob\n"), 0644)

	bin := buildCLI(t)
	out, err := exec.Command(bin, "stats", path).Output()
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, "Rows: 2") {
		t.Errorf("rows: %q", s)
	}
	if !strings.Contains(s, "Cols: 2") {
		t.Errorf("cols: %q", s)
	}
}

func TestCLIJoin(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.csv")
	b := filepath.Join(dir, "b.csv")
	os.WriteFile(a, []byte("id,name\n1,Alice\n2,Bob\n"), 0644)
	os.WriteFile(b, []byte("id,amount\n1,100\n2,200\n"), 0644)

	bin := buildCLI(t)
	out, err := exec.Command(bin, "join", "-on", "id", a, b).Output()
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, "Alice") || !strings.Contains(s, "100") {
		t.Errorf("join: %q", s)
	}
}

func buildCLI(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "gocsv")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = "."
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return bin
}

func TestCLISort(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")
	os.WriteFile(path, []byte("id,age\n1,30\n2,25\n3,35\n"), 0644)

	bin := buildCLI(t)
	out, err := exec.Command(bin, "sort", "-c", "age", path).Output()
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if lines[1] != "2,25" {
		t.Errorf("first: %q", lines[1])
	}
}

func TestCLISortDesc(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.csv")
	os.WriteFile(path, []byte("id,age\n1,30\n2,25\n3,35\n"), 0644)

	bin := buildCLI(t)
	out, err := exec.Command(bin, "sort", "-c", "age", "-desc", path).Output()
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if lines[1] != "3,35" {
		t.Errorf("first: %q", lines[1])
	}
}

func TestCLIDiff(t *testing.T) {
	dir := t.TempDir()
	before := filepath.Join(dir, "before.csv")
	after := filepath.Join(dir, "after.csv")
	os.WriteFile(before, []byte("id,name\n1,Alice\n2,Bob\n3,Charlie\n"), 0644)
	os.WriteFile(after, []byte("id,name\n1,Alice\n2,Bobby\n4,Diana\n"), 0644)

	bin := buildCLI(t)
	out, err := exec.Command(bin, "diff", "-on", "id", before, after).Output()
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, "Added: 1") || !strings.Contains(s, "Removed: 1") || !strings.Contains(s, "Modified: 1") {
		t.Errorf("diff counts: %q", s)
	}
	if !strings.Contains(s, "Diana") || !strings.Contains(s, "Charlie") || !strings.Contains(s, "Bobby") {
		t.Errorf("diff content: %q", s)
	}
}

func TestCLIGenStruct(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	os.WriteFile(path, []byte("id,name,score\n1,Alice,95.5\n2,Bob,87.3\n"), 0644)

	bin := buildCLI(t)
	out, err := exec.Command(bin, "gen-struct", "-name", "User", path).Output()
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, "type User struct") {
		t.Errorf("missing struct: %q", s)
	}
	if !strings.Contains(s, `csv:"id"`) || !strings.Contains(s, `csv:"name"`) || !strings.Contains(s, `csv:"score"`) {
		t.Errorf("missing tags: %q", s)
	}
	if !strings.Contains(s, "int64") || !strings.Contains(s, "string") || !strings.Contains(s, "float64") {
		t.Errorf("types: %q", s)
	}
}

func TestCLIGenStructPackage(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	os.WriteFile(path, []byte("id,name\n1,Alice\n"), 0644)

	bin := buildCLI(t)
	out, err := exec.Command(bin, "gen-struct", "-package", "models", path).Output()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(out), "package models") {
		t.Errorf("package: %q", out)
	}
}

func TestCLISQL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	os.WriteFile(path, []byte("name,age\nAlice,30\nBob,25\nCharlie,35\n"), 0644)

	bin := buildCLI(t)
	out, err := exec.Command(bin, "sql", "SELECT name FROM t WHERE age >= 30 ORDER BY age DESC", path).Output()
	if err != nil {
		t.Fatal(err)
	}
	s := strings.TrimSpace(string(out))
	lines := strings.Split(s, "\n")
	if len(lines) != 3 {
		t.Errorf("lines: %v", lines)
	}
	if lines[1] != "Charlie" || lines[2] != "Alice" {
		t.Errorf("order: %v", lines)
	}
}

func TestCLISQLAggregate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.csv")
	os.WriteFile(path, []byte("city,age\nParis,30\nParis,40\nBerlin,25\n"), 0644)

	bin := buildCLI(t)
	out, err := exec.Command(bin, "sql", "SELECT city, COUNT(*), AVG(age) FROM t GROUP BY city", path).Output()
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	if !strings.Contains(s, "Paris,2,35") || !strings.Contains(s, "Berlin,1,25") {
		t.Errorf("aggregates: %q", s)
	}
}
