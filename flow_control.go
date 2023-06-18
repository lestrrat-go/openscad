package openscad

import (
	"fmt"
	"io"
)

// Let is a statement that declares a variable and assigns it a value.
//
// It's a little clunky to use from Go, but you usually do not need to
// use this from Go (because you could directly insert values from Go
// instead of using OpenSCAD variables), so it's not really a priority for us.
type Let struct {
	variables []*Variable
	children  []Stmt
}

func NewLet(variables ...*Variable) *Let {
	return &Let{
		variables: variables,
	}
}

func (l *Let) Body(children ...Stmt) *Let {
	l.children = make([]Stmt, len(children))
	copy(l.children, children)
	return l
}

func (l *Let) Add(stmt ...Stmt) *Let {
	l.children = append(l.children, stmt...)
	return l
}

func (l *Let) EmitExpr(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, `%slet(`, ctx.Indent())
	for i, v := range l.variables {
		if i > 0 {
			fmt.Fprintf(w, `, `)
		}
		if err := emitExpr(ctx, w, v); err != nil {
			return err
		}
	}
	fmt.Fprintf(w, `)`)
	emitChildren(ctx, w, l.children)
	return nil
}

type ForRange struct {
	start, end, increment interface{}
}

func NewForRange(start, end interface{}) *ForRange {
	return &ForRange{
		start: start,
		end:   end,
	}
}

func (fr *ForRange) Increment(incr interface{}) *ForRange {
	fr.increment = incr
	return fr
}

func (fr *ForRange) EmitExpr(ctx *EmitContext, w io.Writer) error {
	fmt.Fprint(w, `[`)
	emitValue(ctx, w, fr.start)
	fmt.Fprint(w, `:`)
	if incr := fr.increment; incr != nil {
		emitValue(ctx, w, incr)
		fmt.Fprint(w, `:`)
	}
	emitValue(ctx, w, fr.end)
	fmt.Fprint(w, `]`)
	return nil
}

type LoopVar struct {
	variable *Variable
	expr     interface{}
}

func NewLoopVar(variable *Variable, expr interface{}) *LoopVar {
	return &LoopVar{
		variable: variable,
		expr:     expr,
	}
}

func (lv *LoopVar) EmitExpr(ctx *EmitContext, w io.Writer) error {
	if err := emitValue(ctx, w, lv.variable); err != nil {
		return err
	}
	fmt.Fprint(w, `=`)
	if err := emitValue(ctx.WithAllowAssignment(false), w, lv.expr); err != nil {
		return err
	}
	return nil
}

type ForExpr struct {
	loopVars []*LoopVar
	expr     interface{}
}

func NewForExpr(loopVars []*LoopVar) *ForExpr {
	return &ForExpr{
		loopVars: loopVars,
	}
}

func (f *ForExpr) Body(expr interface{}) *ForExpr {
	f.expr = expr
	return f
}

func emitForDecl(ctx *EmitContext, w io.Writer, loopVars []*LoopVar) error {
	fmt.Fprint(w, "for (")
	ctx = ctx.WithAllowAssignment(false)
	for i, v := range loopVars {
		if i > 0 {
			fmt.Fprint(w, `, `)
		}
		if err := emitExpr(ctx, w, v); err != nil {
			return err
		}
	}
	fmt.Fprint(w, `)`)
	return nil
}

func (f *ForExpr) EmitExpr(ctx *EmitContext, w io.Writer) error {
	if err := emitForDecl(ctx, w, f.loopVars); err != nil {
		return err
	}

	if err := emitExpr(ctx, w, f.expr); err != nil {
		return err
	}
	return nil
}

// ForBlock represents a for loop block.
//
// In OpenSCAD for can take two distinctive styles: one as an expression, and
// one as a statement. ForBlock the expression, use ForExpr
type ForBlock struct {
	loopVars []*LoopVar
	children []Stmt
}

func NewFor(loopVars []*LoopVar) *ForBlock {
	return &ForBlock{
		loopVars: loopVars,
	}
}

func (f *ForBlock) Body(stmts ...Stmt) *ForBlock {
	f.children = make([]Stmt, len(stmts))
	copy(f.children, stmts)
	return f
}

func (f *ForBlock) Add(stmts ...Stmt) *ForBlock {
	f.children = append(f.children, stmts...)
	return f
}

func (f *ForBlock) EmitStmt(ctx *EmitContext, w io.Writer) error {
	indent := ctx.Indent()
	fmt.Fprint(w, indent)
	if err := emitForDecl(ctx, w, f.loopVars); err != nil {
		return err
	}
	fmt.Fprint(w, `{`)
	emitChildren(ctx, w, f.children)
	fmt.Fprintf(w, "\n%s}", indent)
	return nil
}

type TernaryOp struct {
	condition interface{}
	trueExpr  interface{}
	falseExpr interface{}
}

func NewTernaryOp(condition, trueExpr, falseExpr interface{}) *TernaryOp {
	return &TernaryOp{
		condition: condition,
		trueExpr:  trueExpr,
		falseExpr: falseExpr,
	}
}

func (op *TernaryOp) EmitExpr(ctx *EmitContext, w io.Writer) error {
	ctx = ctx.WithAllowAssignment(false)
	fmt.Fprint(w, `(`)
	if err := emitExpr(ctx, w, op.condition); err != nil {
		return err
	}
	fmt.Fprint(w, `?`)
	if err := emitExpr(ctx, w, op.trueExpr); err != nil {
		return err
	}
	fmt.Fprint(w, `:`)
	if err := emitExpr(ctx, w, op.falseExpr); err != nil {
		return err
	}
	fmt.Fprint(w, `)`)
	return nil
}

func (op *TernaryOp) EmitStmt(ctx *EmitContext, w io.Writer) error {
	fmt.Fprint(w, ctx.Indent())
	if err := op.EmitExpr(ctx, w); err != nil {
		return err
	}
	fmt.Fprint(w, `;`)
	return nil
}
