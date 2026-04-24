package main

import (
	"fmt"
	"os"
)

var commands = map[string]func(args []string) error{
	"head":       cmdHead,
	"tail":       cmdTail,
	"select":     cmdSelect,
	"filter":     cmdFilter,
	"convert":    cmdConvert,
	"stats":      cmdStats,
	"validate":   cmdValidate,
	"join":       cmdJoin,
	"gen-struct": cmdGenStruct,
	"sql":        cmdSQL,
	"diff":       cmdDiff,
	"sort":       cmdSort,
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	cmd := os.Args[1]
	if cmd == "-h" || cmd == "--help" || cmd == "help" {
		usage()
		return
	}
	fn, ok := commands[cmd]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		usage()
		os.Exit(1)
	}
	if err := fn(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("gocsv — pure Go CSV tooling")
	fmt.Println()
	fmt.Println("Usage: gocsv <command> [flags] <args>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  head        Print first N rows")
	fmt.Println("  tail        Print last N rows")
	fmt.Println("  select      Output only selected columns")
	fmt.Println("  filter      Filter rows by expression")
	fmt.Println("  sort        Sort rows by column")
	fmt.Println("  convert     Convert between CSV/ODS/gzip")
	fmt.Println("  stats       Print row/column/type stats")
	fmt.Println("  validate    Validate against a schema")
	fmt.Println("  join        Join two CSV files on a key")
	fmt.Println("  diff        Diff two CSV files")
	fmt.Println("  sql         Run SQL-like query: SELECT/WHERE/GROUP BY/ORDER BY/LIMIT")
	fmt.Println("  gen-struct  Infer Go struct from CSV columns")
	fmt.Println()
	fmt.Println("Run 'gocsv <command> -h' for command-specific help.")
}
