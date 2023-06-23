package ast

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/lestrrat-go/blackmagic"
)

// Expr represents an expression in the OpenSCAD language.
//
// An arbitrary object may be either an Expr or a Stmt, or both.
// For example, a *Variable is both an Expr and a Stmt.
type Expr interface {
	EmitExpr(*EmitContext, io.Writer) error
}

// Stmt repreents an OpenSCAD statement.
//
// An arbitrary object may be either an Expr or a Stmt, or both.
// For example, a *Variable is both an Expr and a Stmt.
type Stmt interface {
	EmitStmt(*EmitContext, io.Writer) error
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

func (stmts *Stmts) Add(stmt Stmt) {
	*stmts = append(*stmts, stmt)
}

func (stmts Stmts) EmitStmt(ctx *EmitContext, w io.Writer) error {
	ctx = ctx.WithAllowAssignment(true)
	for _, stmt := range stmts {
		if err := stmt.EmitStmt(ctx, w); err != nil {
			return err
		}
	}
	return nil
}

type Declare struct {
	v *Variable
}

func NewDeclare(v *Variable) *Declare {
	return &Declare{
		v: v,
	}
}

func (d *Declare) EmitExpr(ctx *EmitContext, w io.Writer) error {
	return d.v.EmitExpr(ctx.WithAllowAssignment(true), w)
}

func (d *Declare) EmitStmt(ctx *EmitContext, w io.Writer) error {
	return d.v.EmitStmt(ctx, w)
}

// Variable represents a variable in the OpenSCAD language.
// It can be assigned a value so that in appropriate contexts,
// an assignment statement is emitted.
type Variable struct {
	name  string
	value interface{}
}

func NewVariable(name string) *Variable {
	return &Variable{
		name: name,
	}
}

func (p *Variable) String() string {
	var sb strings.Builder
	if err := p.EmitExpr(newEmitContext(), &sb); err != nil {
		panic(err)
	}
	return sb.String()
}

func (p *Variable) HasValue() bool {
	return p.value != nil
}

func (p *Variable) Value(v interface{}) *Variable {
	p.value = v
	return p
}

func (p *Variable) emit(ctx *EmitContext, w io.Writer, isStmt bool) error {
	// This may seem a bit weird, but we want to change the formatting
	// based on if this is a declaration+assignment, or parameter in a
	// function call. The variable is used as statement in the former,
	// and an expression in later
	spacing := ""
	if isStmt {
		spacing = " "
	}
	if ctx.AllowAssignment() && p.value != nil {
		fmt.Fprintf(w, `%[1]s%[2]s=%[2]s`, p.name, spacing)
		// Remove the assignment flag
		if err := emitExpr(ctx.WithAllowAssignment(false), w, p.value); err != nil {
			return err
		}
		return nil
	}
	fmt.Fprintf(w, `%s`, p.name)
	return nil
}

func (p *Variable) EmitExpr(ctx *EmitContext, w io.Writer) error {
	p.emit(ctx, w, false)
	return nil
}

func (p *Variable) EmitStmt(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, "\n%s", ctx.Indent())
	if err := p.emit(ctx.WithAllowAssignment(true), w, true); err != nil {
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

func emitChildren(ctx *EmitContext, w io.Writer, children []Stmt, forceBrace bool) error {
	indent := ctx.Indent()
	numc := len(children)
	if numc == 0 {
		return fmt.Errorf(`expected at least one child`)
	}

	ctx = ctx.IncrIndent()
	if numc == 1 && !forceBrace {
		return children[0].EmitStmt(ctx, w)
	}

	fmt.Fprintf(w, "\n%s{", indent)
	prev := reflect.TypeOf(children[0]).Elem()
	for _, c := range children {
		cur := reflect.TypeOf(c).Elem()
		if cur.Name() != prev.Name() {
			fmt.Fprintf(w, "\n")
		}
		prev = cur
		if err := c.EmitStmt(ctx, w); err != nil {
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

func (m *Module) Add(child Stmt) *Module {
	m.children = append(m.children, child)
	return m
}

func (m *Module) Body(children ...Stmt) *Module {
	m.children = make([]Stmt, len(children))
	copy(m.children, children)
	return m
}

func (m *Module) EmitStmt(ctx *EmitContext, w io.Writer) error {
	indent := ctx.Indent()
	fmt.Fprintf(w, "\n%smodule %s(", indent, m.name)

	{
		pctx := ctx.WithAllowAssignment(true)
		for i, param := range m.parameters {
			if i > 0 {
				fmt.Fprintf(w, ", ")
			}
			emitValue(pctx, w, param)
		}
	}
	fmt.Fprintf(w, ")")

	if err := emitChildren(ctx, w, m.children, true); err != nil {
		return err
	}
	fmt.Fprint(w, "\n")
	return nil
}

type Call struct {
	name       string
	parameters []interface{}
	children   []Stmt
}

func NewCall(name string) *Call {
	return &Call{
		name: name,
	}
}

func (c *Call) String() string {
	var sb strings.Builder
	if err := c.EmitExpr(newEmitContext(), &sb); err != nil {
		panic(err)
	}
	return sb.String()
}

func (c *Call) Parameters(params ...interface{}) *Call {
	c.parameters = append(c.parameters, params...)
	return c
}

func (c *Call) Add(children ...Stmt) *Call {
	c.children = append(c.children, children...)
	return c
}

func (c *Call) EmitStmt(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, "\n%s", ctx.Indent())
	if err := c.EmitExpr(ctx, w); err != nil {
		return err
	}

	if children := c.children; len(children) > 0 {
		return emitChildren(ctx, w, children, false)
	}

	// only emit the last semicolon if there are no children
	fmt.Fprint(w, `;`)
	return nil
}

func (c *Call) EmitExpr(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, `%s(`, c.name)

	ctx = ctx.WithAllowAssignment(true)
	for i, p := range c.parameters {
		if i > 0 {
			fmt.Fprintf(w, `, `)
		}
		if err := emitExpr(ctx, w, p); err != nil {
			return err
		}
	}
	fmt.Fprintf(w, ")")
	return nil
}

type inclusionDirective struct {
	typ  string
	name string
}

func (i *inclusionDirective) EmitStmt(ctx *EmitContext, w io.Writer) error {
	if ctx.Amalgamate() {
		if _, ok := ctx.amalgamated[i.name]; ok {
			return nil
		}

		ctx.amalgamated[i.name] = struct{}{}
		stmts, ok := ctx.registry.Lookup(i.name)
		if !ok {
			return fmt.Errorf(`source file %q not found`, i.name)
		}

		fmt.Fprintf(w, "\n\n// START %s %s\n", i.typ, i.name)
		if err := stmts.EmitStmt(ctx, w); err != nil {
			return err
		}
		fmt.Fprintf(w, "\n// END %s %s\n", i.typ, i.name)
		return nil
	}

	fmt.Fprintf(w, `%s <%s>`, i.typ, i.name)
	return nil
}

type Include struct {
	inclusionDirective
}

func NewInclude(name string) *Include {
	return &Include{
		inclusionDirective: inclusionDirective{
			typ:  `include`,
			name: name,
		},
	}
}

type Use struct {
	inclusionDirective
}

func NewUse(name string) *Use {
	return &Use{
		inclusionDirective: inclusionDirective{
			typ:  `use`,
			name: name,
		},
	}
}

type Index struct {
	expr  interface{}
	index interface{}
}

func NewIndex(expr, index interface{}) *Index {
	return &Index{
		expr:  expr,
		index: index,
	}
}

func (i *Index) String() string {
	var sb strings.Builder
	if err := i.EmitExpr(newEmitContext(), &sb); err != nil {
		panic(err)
	}
	return sb.String()

}

func (i *Index) EmitExpr(ctx *EmitContext, w io.Writer) error {
	if err := emitExpr(ctx, w, i.expr); err != nil {
		return err
	}
	fmt.Fprintf(w, "[")
	if err := emitExpr(ctx, w, i.index); err != nil {
		return err
	}
	fmt.Fprintf(w, "]")
	return nil
}

type BareBlock struct {
	children []Stmt
}

func NewBareBlock(children ...Stmt) *BareBlock {
	return &BareBlock{
		children: children,
	}
}

func (b *BareBlock) Add(children ...Stmt) *BareBlock {
	b.children = append(b.children, children...)
	return b
}

func (b *BareBlock) EmitStmt(ctx *EmitContext, w io.Writer) error {
	return emitChildren(ctx, w, b.children, true)
}
