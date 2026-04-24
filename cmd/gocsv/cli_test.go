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
