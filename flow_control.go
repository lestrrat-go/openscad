package openscad

import (
	"context"
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

func (l *Let) Add(stmt ...Stmt) *Let {
	l.children = append(l.children, stmt...)
	return l
}

func (l *Let) EmitStmt(ctx context.Context, w io.Writer) error {
	indent := GetIndent(ctx)
	fmt.Fprintf(w, `%slet(`, indent)
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

func (fr *ForRange) EmitExpr(ctx context.Context, w io.Writer) error {
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

func (lv *LoopVar) EmitExpr(ctx context.Context, w io.Writer) error {
	if err := emitValue(ctx, w, lv.variable); err != nil {
		return err
	}
	fmt.Fprint(w, `=`)
	if err := emitValue(context.WithValue(ctx, identAssignment{}, false), w, lv.expr); err != nil {
		return err
	}
	return nil
}

type For struct {
	loopVars []*LoopVar
	children []Stmt
}

func NewFor(loopVars []*LoopVar) *For {
	return &For{
		loopVars: loopVars,
	}
}

func (f *For) Add(stmts ...Stmt) *For {
	f.children = append(f.children, stmts...)
	return f
}

func (f *For) EmitStmt(ctx context.Context, w io.Writer) error {
	indent := GetIndent(ctx)
	fmt.Fprintf(w, "%sfor(", indent)
	lctx := context.WithValue(ctx, identAssignment{}, false)
	for i, v := range f.loopVars {
		if i > 0 {
			fmt.Fprint(w, `, `)
		}
		if err := emitValue(lctx, w, v); err != nil {
			return err
		}
	}
	fmt.Fprint(w, `) {`)
	emitChildren(ctx, w, f.children)
	fmt.Fprintf(w, "\n%s}", indent)
	return nil
}
