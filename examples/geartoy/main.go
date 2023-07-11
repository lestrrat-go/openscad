//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/lestrrat-go/openscad/ast"
	_ "github.com/lestrrat-go/openscad/examples/geartoy"
)

// Executing this as `go run main.go` will output the amalgamated OpenSCAD
// code to generate the design.
func main() {
	if err := _main(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func _main() error {
	return ast.EmitFile(`geartoy.scad`, os.Stdout, ast.WithAmalgamation())
}
