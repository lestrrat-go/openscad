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

func (t *Translate) EmitExpr(ctx *EmitContext, w io.Writer) error {
	fmt.Fprint(w, `translate(`)
	if err := emitExpr(ctx.WithAllowAssignment(false), w, t.v); err != nil {
		return fmt.Errorf(`failed to emit translate vector: %v`, err)
	}
	fmt.Fprint(w, `)`)
	return emitChildren(ctx, w, t.children, true)
}

func (t *Translate) EmitStmt(ctx *EmitContext, w io.Writer) error {
	fmt.Fprint(w, ctx.Indent())
	return t.EmitExpr(ctx, w)
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
	indent := ctx.Indent()
	fmt.Fprintf(w, `%srotate(`, indent)
	if err := emitExpr(ctx, w, r.v); err != nil {
		return fmt.Errorf(`failed to emit rotate vector: %w`, err)
	}
	fmt.Fprint(w, `)`)

	return emitChildren(ctx, w, r.children, false)
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

func (l *LinearExtrude) Add(s Stmt) *LinearExtrude {
	l.children = append(l.children, s)
	return l
}

func (l *LinearExtrude) Fn(fn int) *LinearExtrude {
	l.fn = &fn
	return l
}

func (l *LinearExtrude) EmitStmt(ctx *EmitContext, w io.Writer) error {
	ctx = ctx.WithAllowAssignment(false)
	fmt.Fprintf(w, `%slinear_extrude(height=`, ctx.Indent())
	if l.height == nil {
		return fmt.Errorf("height must be specified")
	}
	if err := emitValue(ctx, w, l.height); err != nil {
		return fmt.Errorf(`failed to emit linear_extrude height: %w`, err)
	}

	if l.center != nil {
		fmt.Fprint(w, `, center=`)
		if err := emitValue(ctx, w, l.center); err != nil {
			return fmt.Errorf(`failed to emit linear_extrude center: %w`, err)
		}
	}
	if l.convexity != nil {
		fmt.Fprintf(w, `, convexity=`)
		if err := emitValue(ctx, w, l.convexity); err != nil {
			return fmt.Errorf(`failed to emit linear_extrude convexity: %w`, err)
		}
	}
	if l.twist != nil {
		fmt.Fprintf(w, `, twist=`)
		if err := emitValue(ctx, w, l.twist); err != nil {
			return fmt.Errorf(`failed to emit linear_extrude twist: %w`, err)
		}
	}
	if l.scale != nil {
		fmt.Fprintf(w, `, scale=`)
		if err := emitValue(ctx, w, l.scale); err != nil {
			return fmt.Errorf(`failed to emit linear_extrude scale: %w`, err)
		}
	}
	if l.fn != nil {
		emitFn(w, l.fn)
	}
	fmt.Fprint(w, `)`)
	return emitChildren(ctx, w, l.children, false)
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
