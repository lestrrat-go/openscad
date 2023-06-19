package openscad

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/lestrrat-go/openscad/ast"
)

func Register(name string, stmt ast.Stmt) error {
	return ast.Register(name, stmt)
}

func Lookup(name string) (ast.Stmt, bool) {
	return ast.Lookup(name)
}

func RegisterFile(filename string, options ...RegisterFileOption) error {
	srcfs := os.DirFS(".")
	lookupName := filename

	//nolint:forcetypeassert
	for _, option := range options {
		switch option.Ident() {
		case optLookupNameKey{}:
			lookupName = option.Value().(string)
		case optFSKey{}:
			srcfs = option.Value().(fs.FS)
		}
	}

	code, err := fs.ReadFile(srcfs, filename)
	if err != nil {
		return fmt.Errorf("failed to read %q: %w", filename, err)
	}

	stmts, err := Parse(code)
	if err != nil {
		return fmt.Errorf("failed to parse %q: %w", filename, err)
	}

	return ast.Register(lookupName, stmts)
}
