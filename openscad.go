package openscad

import (
	"context"
	"fmt"
	"io"

	"github.com/lestrrat-go/blackmagic"
)

// Expr represents an expression in the OpenSCAD language.
//
// An arbitrary object may be either an Expr or a Stmt, or both.
// For example, a *Variable is both an Expr and a Stmt.
type Expr interface {
	EmitExpr(context.Context, io.Writer) error
}

// Stmt repreents an OpenSCAD statement.
//
// An arbitrary object may be either an Expr or a Stmt, or both.
// For example, a *Variable is both an Expr and a Stmt.
type Stmt interface {
	EmitStmt(context.Context, io.Writer) error
}

type identFa struct{}
type identFn struct{}
type identFs struct{}
type identIndent struct{}
type identAssignment struct{}

func IdentFn() interface{} {
	return identFn{}
}

func GetValue(ctx context.Context, ident interface{}, ptr interface{}) error {
	val := ctx.Value(ident)
	return blackmagic.AssignIfCompatible(ptr, val)
}

func GetIndent(ctx context.Context) string {
	s := ctx.Value(identIndent{})
	if s == nil {
		return ""
	}
	if str, ok := s.(string); ok {
		return str
	}
	return ""
}

func AddIndent(ctx context.Context) context.Context {
	return context.WithValue(ctx, identIndent{}, GetIndent(ctx)+"  ")
}

// Stmts is a sequence of statements.
type Stmts []Stmt

func (stmts Stmts) Emit(ctx context.Context, w io.Writer) error {
	return stmts.EmitStmt(ctx, w)
}

func (stmts Stmts) EmitStmt(ctx context.Context, w io.Writer) error {
	ctx = context.WithValue(ctx, identAssignment{}, true)
	for i, stmt := range stmts {
		if i > 0 {
			fmt.Fprintf(w, "\n")
		}
		if err := stmt.EmitStmt(ctx, w); err != nil {
			return err
		}
	}
	return nil
}

type Shape interface {
	Emit(context.Context, io.Writer) error
}

type emitter interface {
	Emit(context.Context, io.Writer) error
}

func emitValue(ctx context.Context, w io.Writer, v interface{}) error {
	switch v := v.(type) {
	case Expr:
		return v.EmitExpr(ctx, w)
	case Stmt:
		return v.EmitStmt(ctx, w)
	default:
		_, err := fmt.Fprintf(w, "%#v", v)
		return err
	}
}

type Variable struct {
	name  string
	value interface{}
}

func NewVariable(name string) *Variable {
	return &Variable{
		name: name,
	}
}

func (p *Variable) HasValue() bool {
	return p.value != nil
}

func (p *Variable) Value(v interface{}) *Variable {
	p.value = v
	return p
}

func (p *Variable) EmitExpr(ctx context.Context, w io.Writer) error {
	if getBool(ctx, identAssignment{}) {
		fmt.Fprintf(w, `%s=`, p.name)
		// Remove the assignment flag
		if err := emitValue(context.WithValue(ctx, identAssignment{}, false), w, p.value); err != nil {
			return err
		}
		return nil
	}
	fmt.Fprintf(w, `%s`, p.name)
	return nil
}

func (p *Variable) EmitStmt(ctx context.Context, w io.Writer) error {
	if err := p.EmitExpr(ctx, w); err != nil {
		return err
	}
	fmt.Fprint(w, `;`)
	return nil
}

func getBool(ctx context.Context, ident interface{}) bool {
	v := ctx.Value(ident)
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return false
}

type Module struct {
	name       string
	parameters []*Variable
	children   []Stmt
}

func emitChildren(ctx context.Context, w io.Writer, children []Stmt) error {
	indent := GetIndent(ctx)
	numc := len(children)
	if numc == 0 {
		return fmt.Errorf(`expected at least one child`)
	}

	if numc == 1 {
		fmt.Fprintf(w, "\n")
		return children[0].EmitStmt(AddIndent(ctx), w)
	}

	fmt.Fprintf(w, "\n%s{", indent)
	for _, c := range children {
		fmt.Fprintf(w, "\n")
		if err := c.EmitStmt(AddIndent(ctx), w); err != nil {
			return err
		}
	}
	fmt.Fprintf(w, "\n%s}", indent)
	return nil
}

func NewModule(name string) *Module {
	return &Module{
		name: name,
	}
}

func (m *Module) Parameters(params ...*Variable) *Module {
	m.parameters = append(m.parameters, params...)
	return m
}

func (m *Module) Actions(children ...Stmt) *Module {
	m.children = append(m.children, children...)
	return m
}

func (m *Module) EmitStmt(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, `module %s(`, m.name)

	{
		pctx := context.WithValue(ctx, identAssignment{}, true)
		for i, param := range m.parameters {
			if i > 0 {
				fmt.Fprintf(w, ", ")
			}
			emitValue(pctx, w, param)
		}
	}
	fmt.Fprintf(w, ")\n{")

	for _, c := range m.children {
		fmt.Fprintf(w, "\n")
		if err := c.EmitStmt(AddIndent(ctx), w); err != nil {
			return err
		}
	}
	fmt.Fprintf(w, "\n}")
	return nil
}

type Call struct {
	name       string
	parameters []*Variable
}

func NewCall(name string) *Call {
	return &Call{
		name: name,
	}
}

func (c *Call) Parameters(params ...*Variable) *Call {
	c.parameters = append(c.parameters, params...)
	return c
}

func (c *Call) EmitStmt(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, `%s(`, c.name)

	{
		pctx := context.WithValue(ctx, identAssignment{}, true)
		for i, param := range c.parameters {
			if i > 0 {
				fmt.Fprintf(w, ", ")
			}
			emitValue(pctx, w, param)
		}
	}
	fmt.Fprintf(w, ");")
	return nil
}
