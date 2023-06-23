package ast

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	asNone = iota
	asExpr
	asStmt
)

// EmitContext holds the context for emitting OpenSCAD code.
// The object is immutable once created. To change the values,
// create a new context using one of the provided methods
type EmitContext struct {
	amalgamated     map[string]struct{}
	registry        *Registry
	indent          string
	as              int
	allowAssignment bool
	amalgamate      bool
}

func newEmitContext() *EmitContext {
	return &EmitContext{
		allowAssignment: true,
		registry:        globalRegistry,
	}
}

func (e *EmitContext) Amalgamate() bool {
	return e.amalgamate
}

func (e *EmitContext) Copy() *EmitContext {
	return &EmitContext{
		indent:          e.indent,
		amalgamate:      e.amalgamate,
		amalgamated:     e.amalgamated,
		registry:        e.registry,
		as:              e.as,
		allowAssignment: e.allowAssignment,
	}
}

func (e *EmitContext) ForceExpr() *EmitContext {
	e2 := e.Copy()
	e2.as = asExpr
	return e2
}

func (e *EmitContext) ForceStmt() *EmitContext {
	e2 := e.Copy()
	e2.as = asStmt
	return e2
}

func (e *EmitContext) AsExpr() bool {
	return e.as == asExpr
}

func (e *EmitContext) AsStmt() bool {
	return e.as == asStmt
}

func (e *EmitContext) AllowAssignment() bool {
	return e.allowAssignment
}

func (e *EmitContext) Indent() string {
	return e.indent
}

func (e *EmitContext) WithIndent(indent string) *EmitContext {
	e2 := e.Copy()
	e2.indent = indent
	return e2
}

func (e *EmitContext) WithAllowAssignment(allowAssignment bool) *EmitContext {
	e2 := e.Copy()
	e2.allowAssignment = allowAssignment
	return e2
}

const singleIndent = "  "

func (e *EmitContext) IncrIndent() *EmitContext {
	return e.WithIndent(e.indent + singleIndent)
}

func (e *EmitContext) DecrIndent() *EmitContext {
	if e.indent == "" {
		return e
	}
	if len(e.indent) < len(singleIndent) {
		return e.WithIndent("")
	}
	return e.WithIndent(e.indent[:len(e.indent)-len(singleIndent)])
}

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
		return fmt.Errorf(`failed to execute EmitFile: no such file: %s`, filename)
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

func EmitString(stmt Stmt, options ...EmitOption) (string, error) {
	var buf bytes.Buffer
	if err := Emit(stmt, &buf, options...); err != nil {
		return "", err
	}
	return buf.String(), nil
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

func emitStmt(ctx *EmitContext, w io.Writer, v interface{}) error {
	return emitValue(ctx.ForceStmt(), w, v)
}

func emitListContent(ctx *EmitContext, w io.Writer, rv reflect.Value) error {
	// lists are expressed as a single line if they contain 3 or fewer elements
	separateLine := rv.Len() > 3
	var body bytes.Buffer
	for i := 0; i < rv.Len(); i++ {
		if i > 0 {
			fmt.Fprintf(&body, ", ")
		}
		if separateLine {
			fmt.Fprintln(&body)
		}
		if err := emitValue(ctx, &body, rv.Index(i).Interface()); err != nil {
			return err
		}
	}

	if _, err := body.WriteTo(w); err != nil {
		return fmt.Errorf(`failed to write list content: %w`, err)
	}
	return nil
}

func emitAny(ctx *EmitContext, w io.Writer, v interface{}) error {
	childIndent := ctx.Indent() + singleIndent
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice:
		fmt.Fprint(w, "[")

		var content bytes.Buffer
		if err := emitListContent(ctx, &content, rv); err != nil {
			return err
		}

		// if the result contains any newlines, we emit it as a block
		if strings.ContainsRune(content.String(), '\n') {
			if err := addIndent(w, &content, childIndent); err != nil {
				return err
			}
			fmt.Fprintf(w, "\n%s]", ctx.Indent())
		} else {
			if _, err := content.WriteTo(w); err != nil {
				return fmt.Errorf(`failed to write list content: %w`, err)
			}
			fmt.Fprint(w, "]")
		}

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
