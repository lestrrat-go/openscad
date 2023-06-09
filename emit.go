package openscad

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
)

func EmitFile(filename string, w io.Writer, options ...EmitFileOption) error {
	registry := globalRegistry

	emitOptions := make([]EmitOption, 0, len(options))
	emitOptions = append(emitOptions, WithRegistry(registry))
	//nolint:forcetypeassert
	for _, option := range options {
		switch option.Ident() {
		case optRegistryKey{}:
			registry = option.Value().(*Registry)
		case optAmalgamationKey{}:
			emitOptions = append(emitOptions, option)
		}
	}

	stmt, ok := registry.Lookup(filename)
	if !ok {
		return fmt.Errorf(`no such file: %s`, filename)
	}

	if err := Emit(stmt, w, emitOptions...); err != nil {
		return err
	}
	return nil
}

func WriteFile(filename string, options ...WriteFileOption) error {
	var dir string

	emitFileOptions := make([]EmitFileOption, 0, len(options))
	//nolint:forcetypeassert
	for _, option := range options {
		switch option.Ident() {
		case optOutputDirKey{}:
			dir = option.Value().(string)
		default:
			switch option := option.(type) {
			case EmitFileOption:
				emitFileOptions = append(emitFileOptions, option)
			}
		}
	}
	var buf bytes.Buffer
	if err := EmitFile(filename, &buf, emitFileOptions...); err != nil {
		return err
	}
	f, err := os.CreateTemp("", "go-openscad-*")
	if err != nil {
		return err
	}
	if _, err := buf.WriteTo(f); err != nil {
		return err
	}
	f.Close()
	defer os.Remove(f.Name())

	if dir != "" {
		filename = filepath.Join(dir, filename)
	}
	if err := os.Rename(f.Name(), filename); err != nil {
		return err
	}
	return nil
}

// Emit takes a statement (or a list of statements) and emits them
// into the writer.
//
// By default it emits the statements as a regular OpenSCAD file,
// but you can ask it to create an "amalgamated" file by passing
// the WithAmalgamate option.
func Emit(stmt Stmt, w io.Writer, options ...EmitOption) error {
	ctx := newEmitContext()

	//nolint:forcetypeassert
	for _, option := range options {
		switch option.Ident() {
		case optAmalgamationKey{}:
			ctx.amalgamate = option.Value().(bool)
			if ctx.amalgamate {
				ctx.amalgamated = make(map[string]struct{})
			}
		case optRegistryKey{}:
			ctx.registry = option.Value().(*Registry)
		}
	}
	return stmt.EmitStmt(ctx, w)
}

func emitExpr(ctx *EmitContext, w io.Writer, v interface{}) error {
	return emitValue(ctx.ForceExpr(), w, v)
}

func emitAny(ctx *EmitContext, w io.Writer, v interface{}) error {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice:
		fmt.Fprint(w, "[")
		for i := 0; i < rv.Len(); i++ {
			if i > 0 {
				fmt.Fprintf(w, ", ")
			}
			if err := emitValue(ctx, w, rv.Index(i).Interface()); err != nil {
				return err
			}
		}
		fmt.Fprint(w, "]")
	default:
		_, err := fmt.Fprintf(w, "%#v", v)
		if err != nil {
			return err
		}
	}
	return nil
}

func emitValue(ctx *EmitContext, w io.Writer, v interface{}) error {
	if ctx.AsExpr() {
		if e, ok := v.(Expr); ok {
			return e.EmitExpr(ctx, w)
		}
	} else if ctx.AsStmt() {
		if e, ok := v.(Stmt); ok {
			return e.EmitStmt(ctx, w)
		}
	} else {
		switch v := v.(type) {
		case Expr:
			return v.EmitExpr(ctx, w)
		case Stmt:
			return v.EmitStmt(ctx, w)
		}
	}

	return emitAny(ctx, w, v)
}
