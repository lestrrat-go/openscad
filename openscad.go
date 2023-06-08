package openscad

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/lestrrat-go/blackmagic"
)

var globalRegistry = &Registry{
	storage: make(map[string]Stmt),
}

func Register(name string, s Stmt) error {
	return globalRegistry.Register(name, s)
}

func Lookup(name string) (Stmt, bool) {
	return globalRegistry.Lookup(name)
}

type Registry struct {
	mu      sync.RWMutex
	storage map[string]Stmt
}

func (r *Registry) Register(name string, s Stmt) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.storage[name] = s
	return nil
}

func (r *Registry) Lookup(name string) (Stmt, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.storage == nil {
		return nil, false
	}
	s, ok := r.storage[name]
	return s, ok
}

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
	if getBool(ctx, identAssignment{}) && p.value != nil {
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
	fmt.Fprint(w, GetIndent(ctx))
	if err := p.EmitExpr(context.WithValue(ctx, identAssignment{}, true), w); err != nil {
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
	fmt.Fprintf(w, "\nmodule %s(", m.name)

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

func (c *Call) EmitStmt(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, `%s`, GetIndent(ctx))
	if err := c.EmitExpr(ctx, w); err != nil {
		return err
	}

	if children := c.children; len(children) > 0 {
		return emitChildren(ctx, w, children)
	}
	fmt.Fprint(w, `;`)
	return nil
}

func (c *Call) EmitExpr(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, `%s(`, c.name)

	ctx = context.WithValue(ctx, identAssignment{}, false)
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

func (f *Function) Parameters(params ...*Variable) *Function {
	f.parameters = append(f.parameters, params...)
	return f
}

func (f *Function) Body(body interface{}) *Function {
	f.body = body
	return f
}

func (f *Function) EmitStmt(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, `%s`, GetIndent(ctx))
	if err := f.EmitExpr(ctx, w); err != nil {
		return err
	}
	fmt.Fprint(w, `;`)
	return nil
}

func (f *Function) EmitExpr(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, `function %s(`, f.name)

	ctx = context.WithValue(ctx, identAssignment{}, false)
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
	return emitExpr(ctx, w, f.body)
}

type Include struct {
	name string
}

func NewInclude(name string) *Include {
	return &Include{
		name: name,
	}
}

func (i *Include) EmitStmt(_ context.Context, w io.Writer) error {
	fmt.Fprintf(w, `include <%s>`, i.name)
	return nil
}

type Use struct {
	name string
}

func NewUse(name string) *Use {
	return &Use{
		name: name,
	}
}

func (i *Use) EmitStmt(_ context.Context, w io.Writer) error {
	fmt.Fprintf(w, `use <%s>`, i.name)
	return nil
}
