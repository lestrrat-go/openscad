package openscad

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
	emitExpr(ctx.WithAllowAssignment(false), w, t.v)
	fmt.Fprint(w, `)`)
	return emitChildren(ctx, w, t.children)
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
	emitExpr(ctx, w, r.v)
	fmt.Fprint(w, `)`)

	return emitChildren(ctx, w, r.children)
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
	emitValue(ctx, w, l.height)
	if l.center != nil {
		fmt.Fprint(w, `, center=`)
		emitValue(ctx, w, l.center)
	}
	if l.convexity != nil {
		fmt.Fprintf(w, `, convexity=`)
		emitValue(ctx, w, l.convexity)
	}
	if l.twist != nil {
		fmt.Fprintf(w, `, twist=`)
		emitValue(ctx, w, l.twist)
	}
	if l.scale != nil {
		fmt.Fprintf(w, `, scale=`)
		emitValue(ctx, w, l.scale)
	}
	if l.fn != nil {
		emitFn(w, l.fn)
	}
	fmt.Fprint(w, `)`)
	return emitChildren(ctx, w, l.children)
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
