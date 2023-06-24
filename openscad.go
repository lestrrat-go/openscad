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
	lookupName := filename

	var parseFileOptions []ParseFileOption
	//nolint:forcetypeassert
	for _, option := range options {
		switch option.Ident() {
		case optLookupNameKey{}:
			lookupName = option.Value().(string)
		default:
			if pfo, ok := option.(ParseFileOption); ok {
				parseFileOptions = append(parseFileOptions, pfo)
			}
		}
	}

	stmts, err := ParseFile(filename, parseFileOptions...)
	if err != nil {
		return err
	}
	return ast.Register(lookupName, stmts)
}

func ParseFile(filename string, options ...ParseFileOption) (ast.Stmt, error) {
	srcfs := os.DirFS(".")

	//nolint:forcetypeassert
	for _, option := range options {
		switch option.Ident() {
		case optFSKey{}:
			srcfs = option.Value().(fs.FS)
		}
	}

	code, err := fs.ReadFile(srcfs, filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %w", filename, err)
	}

	stmts, err := Parse(code)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %q: %w", filename, err)
	}
	return stmts, nil
}
