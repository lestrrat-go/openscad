package ast

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Function struct {
	name       string
	parameters []*Variable
	body       interface{}
}

func NewFunction(name string) *Function {
	return &Function{
		name: name,
	}
}

func (f *Function) Name() string {
	return f.name
}

func (f *Function) Parameters(params ...*Variable) *Function {
	f.parameters = append(f.parameters, params...)
	return f
}

func (f *Function) Body(body interface{}) *Function {
	f.body = body
	return f
}

func (f *Function) EmitStmt(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, "\n%s", ctx.Indent())
	if err := f.EmitExpr(ctx, w); err != nil {
		return err
	}
	fmt.Fprint(w, ";")
	return nil
}

func (f *Function) EmitExpr(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, `function %s(`, f.name)

	ctx = ctx.WithAllowAssignment(true)
	for i, p := range f.parameters {
		if i > 0 {
			fmt.Fprintf(w, `, `)
		}
		if err := emitExpr(ctx, w, p); err != nil {
			return err
		}
	}
	fmt.Fprintf(w, ") = ")

	if f.body == nil {
		return fmt.Errorf(`expected a body`)
	}

	var body bytes.Buffer
	if err := emitExpr(ctx, &body, f.body); err != nil {
		return err
	}

	if strings.ContainsRune(body.String(), '\n') {
		fmt.Fprint(w, "\n")
		if err := addIndent(w, &body, ctx.Indent()+singleIndent); err != nil {
			return err
		}
	} else {
		if _, err := body.WriteTo(w); err != nil {
			return fmt.Errorf(`failed to write function body: %w`, err)
		}
	}
	return nil
}

type LookupStmt struct {
	key    interface{}
	values interface{}
}

func NewLookup(key, values interface{}) *LookupStmt {
	return &LookupStmt{
		key:    key,
		values: values,
	}
}

func (l *LookupStmt) EmitExpr(ctx *EmitContext, w io.Writer) error {
	fmt.Fprint(w, "lookup(")
	ctx = ctx.WithAllowAssignment(false)

	if err := emitExpr(ctx, w, l.key); err != nil {
		return err
	}
	fmt.Fprintf(w, ", ")

	if err := emitExpr(ctx, w, l.values); err != nil {
		return err
	}
	return nil
}
