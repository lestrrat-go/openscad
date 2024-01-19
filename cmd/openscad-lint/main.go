package main

import (
	"fmt"
	"os"

	"github.com/lestrrat-go/openscad"
	"github.com/lestrrat-go/openscad/ast"
)

func main() {
	if err := _main(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func _main() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: %s <file>", os.Args[0])
	}

	stmts, err := openscad.ParseFile(os.Args[1])
	if err != nil {
		return fmt.Errorf("failed to parse file: %s", err)
	}

	return ast.Emit(stmts, os.Stdout)
}
