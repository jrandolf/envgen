// envgen generates typed config code from .env.schema files.
//
// Usage:
//
//	envgen -lang=go  -schema=path/to/.env.schema -out=config.go -package=config
//	envgen -lang=py  -schema=path/to/.env.schema -out=config.py
//	envgen -lang=ts  -schema=path/to/.env.schema -out=lib/env.ts
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jrandolf/envgen/internal/codegen"
	"github.com/jrandolf/envgen/internal/parser"
)

func main() {
	lang := flag.String("lang", "", "output language: go, py, rs, or ts")
	schema := flag.String("schema", "", "path to .env.schema file")
	out := flag.String("out", "", "output file path")
	pkg := flag.String("package", "config", "Go package name (only for -lang=go)")
	flag.Parse()

	if *lang == "" || *schema == "" || *out == "" {
		fmt.Fprintf(os.Stderr, "usage: envgen -lang=<go|py|rs|ts> -schema=<path> -out=<path> [-package=<name>]\n")
		os.Exit(1)
	}

	s, err := parser.ParseFile(*schema)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse: %v\n", err)
		os.Exit(1)
	}

	// Ensure output directory exists.
	if dir := filepath.Dir(*out); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir: %v\n", err)
			os.Exit(1)
		}
	}

	f, err := os.Create(*out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	switch *lang {
	case "go":
		err = codegen.GenerateGo(f, s, *pkg)
	case "py":
		err = codegen.GeneratePython(f, s)
	case "rs":
		err = codegen.GenerateRust(f, s)
	case "ts":
		err = codegen.GenerateTypeScript(f, s)
	default:
		fmt.Fprintf(os.Stderr, "unsupported language: %s\n", *lang)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "generate: %v\n", err)
		os.Exit(1)
	}
}
