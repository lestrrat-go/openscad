package openscad

import (
	"context"
	"fmt"
	"io"

	"github.com/lestrrat-go/blackmagic"
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
	nestedBinaryOp  bool
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
		nestedBinaryOp:  e.nestedBinaryOp,
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

func (e *EmitContext) IsNestedBinaryOp() bool {
	return e.nestedBinaryOp
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

func (e *EmitContext) WithNestedBinaryOp(v bool) *EmitContext {
	e2 := e.Copy()
	e2.nestedBinaryOp = v
	return e2
}

const indent = "  "

func (e *EmitContext) IncrIndent() *EmitContext {
	return e.WithIndent(e.indent + indent)
}

func (e *EmitContext) DecrIndent() *EmitContext {
	if e.indent == "" {
		return e
	}
	if len(e.indent) < len(indent) {
		return e.WithIndent("")
	}
	return e.WithIndent(e.indent[:len(e.indent)-len(indent)])
}

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

func (stmts Stmts) EmitStmt(ctx *EmitContext, w io.Writer) error {
	ctx = ctx.WithAllowAssignment(true)
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

func (p *Variable) HasValue() bool {
	return p.value != nil
}

func (p *Variable) Value(v interface{}) *Variable {
	p.value = v
	return p
}

func (p *Variable) EmitExpr(ctx *EmitContext, w io.Writer) error {
	if ctx.AllowAssignment() && p.value != nil {
		fmt.Fprintf(w, `%s=`, p.name)
		// Remove the assignment flag
		if err := emitExpr(ctx.WithAllowAssignment(false), w, p.value); err != nil {
			return err
		}
		return nil
	}
	fmt.Fprintf(w, `%s`, p.name)
	return nil
}

func (p *Variable) EmitStmt(ctx *EmitContext, w io.Writer) error {
	fmt.Fprint(w, ctx.Indent())
	if err := p.EmitExpr(ctx.WithAllowAssignment(true), w); err != nil {
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

func emitChildren(ctx *EmitContext, w io.Writer, children []Stmt) error {
	indent := ctx.Indent()
	numc := len(children)
	if numc == 0 {
		return fmt.Errorf(`expected at least one child`)
	}

	ctx = ctx.IncrIndent()
	if numc == 1 {
		fmt.Fprintf(w, "\n")
		return children[0].EmitStmt(ctx, w)
	}

	fmt.Fprintf(w, "\n%s{", indent)
	for _, c := range children {
		fmt.Fprintf(w, "\n")
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

func (m *Module) Body(children ...Stmt) *Module {
	m.children = make([]Stmt, len(children))
	copy(m.children, children)
	return m
}

func (m *Module) EmitStmt(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, "\nmodule %s(", m.name)

	{
		pctx := ctx.WithAllowAssignment(true)
		for i, param := range m.parameters {
			if i > 0 {
				fmt.Fprintf(w, ", ")
			}
			emitValue(pctx, w, param)
		}
	}
	fmt.Fprintf(w, ")\n{")

	ctx = ctx.IncrIndent()
	for _, c := range m.children {
		fmt.Fprintf(w, "\n")
		if err := c.EmitStmt(ctx, w); err != nil {
			return err
		}
	}
	fmt.Fprintf(w, "\n}")
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

func (c *Call) Parameters(params ...interface{}) *Call {
	c.parameters = append(c.parameters, params...)
	return c
}

func (c *Call) Add(children ...Stmt) *Call {
	c.children = append(c.children, children...)
	return c
}

func (c *Call) EmitStmt(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, `%s`, ctx.Indent())
	if err := c.EmitExpr(ctx, w); err != nil {
		return err
	}

	if children := c.children; len(children) > 0 {
		return emitChildren(ctx, w, children)
	}

	// only emit the last semicolon if there are no children
	fmt.Fprint(w, `;`)
	return nil
}

func (c *Call) EmitExpr(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, `%s(`, c.name)

	ctx = ctx.WithAllowAssignment(false)
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

type Render struct{ noArgBlock }

func NewRender() *Render {
	return &Render{noArgBlock{name: `render`}}
}

func (r *Render) Body(children ...Stmt) *Render {
	r.children = make([]Stmt, len(children))
	copy(r.children, children)
	return r
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
