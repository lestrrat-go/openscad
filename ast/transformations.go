package ast

import (
	"fmt"
	"io"
)

type Translate struct {
	v        interface{}
	children []Stmt
}

func NewTranslate(v interface{}, children ...Stmt) *Translate {
	return &Translate{
		v:        v,
		children: children,
	}
}

func (t *Translate) Body(children ...Stmt) *Translate {
	t.children = make([]Stmt, len(children))
	copy(t.children, children)
	return t
}

func (t *Translate) Add(s Stmt) *Translate {
	t.children = append(t.children, s)
	return t
}

func (t *Translate) makeCall() *Call {
	call := NewCall(`translate`).
		Parameters(t.v)
	if children := t.children; len(children) > 0 {
		call.Add(children...)
	}
	return call
}

func (t *Translate) EmitExpr(ctx *EmitContext, w io.Writer) error {
	call := t.makeCall()
	return call.EmitExpr(ctx, w)
}

func (t *Translate) EmitStmt(ctx *EmitContext, w io.Writer) error {
	call := t.makeCall()
	return call.EmitStmt(ctx, w)
}

type Rotate struct {
	v        interface{}
	children []Stmt
}

func NewRotate(v interface{}, children ...Stmt) *Rotate {
	return &Rotate{
		v:        v,
		children: children,
	}
}

func (r *Rotate) Body(children ...Stmt) *Rotate {
	r.children = make([]Stmt, len(children))
	copy(r.children, children)
	return r
}

func (r *Rotate) Add(s Stmt) *Rotate {
	r.children = append(r.children, s)
	return r
}

func (r *Rotate) EmitStmt(ctx *EmitContext, w io.Writer) error {
	call := NewCall(`rotate`).
		Parameters(r.v)
	if children := r.children; len(children) > 0 {
		call.Add(children...)
	}
	return call.EmitStmt(ctx, w)
}

type LinearExtrude struct {
	height    interface{}
	center    interface{}
	convexity interface{}
	twist     interface{}
	scale     interface{}
	fn        *int
	children  []Stmt
}

func NewLinearExtrude(height, center, convexity, twist, scale interface{}, children ...Stmt) *LinearExtrude {
	return &LinearExtrude{
		height:    height,
		center:    center,
		convexity: convexity,
		twist:     twist,
		scale:     scale,
		children:  children,
	}
}

func (l *LinearExtrude) Add(stmts ...Stmt) *LinearExtrude {
	l.children = append(l.children, stmts...)
	return l
}

func (l *LinearExtrude) Fn(fn int) *LinearExtrude {
	l.fn = &fn
	return l
}

func (l *LinearExtrude) EmitStmt(ctx *EmitContext, w io.Writer) error {
	var parameters []interface{}
	if l.height == nil {
		return fmt.Errorf("height must be specified")
	}
	parameters = append(parameters, NewVariable("height").Value(l.height))

	if v := l.center; v != nil {
		parameters = append(parameters, NewVariable("center").Value(v))
	}
	if v := l.convexity; v != nil {
		parameters = append(parameters, NewVariable("convexity").Value(v))
	}
	if v := l.twist; v != nil {
		parameters = append(parameters, NewVariable("twist").Value(v))
	}
	if v := l.scale; v != nil {
		parameters = append(parameters, NewVariable("scale").Value(v))
	}
	if l.fn != nil {
		parameters = append(parameters, NewVariable("$fn").Value(l.fn))
	}
	call := NewCall("linear_extrude").
		Parameters(parameters...)
	if children := l.children; len(children) > 0 {
		call = call.Add(children...)
	}

	return call.EmitStmt(ctx, w)
}

type Hull struct{ noArgBlock }

func NewHull() *Hull {
	return &Hull{noArgBlock{name: "hull"}}
}

func (h *Hull) Add(s Stmt) *Hull {
	h.noArgBlock.Add(s)
	return h
}

func (h *Hull) Body(s ...Stmt) *Hull {
	h.noArgBlock.Body(s...)
	return h
}
