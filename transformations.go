package openscad

import (
	"context"
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

func (t *Translate) Add(s Stmt) *Translate {
	t.children = append(t.children, s)
	return t
}

func (t *Translate) EmitExpr(ctx context.Context, w io.Writer) error {
	fmt.Fprint(w, `translate(`)
	emitExpr(ctx, w, t.v)
	fmt.Fprint(w, `)`)
	return emitChildren(ctx, w, t.children)
}

func (t *Translate) EmitStmt(ctx context.Context, w io.Writer) error {
	fmt.Fprint(w, GetIndent(ctx))
	return t.EmitExpr(ctx, w)
}

type Rotate struct {
	dx, dy, dz interface{} // angles
	children   []Stmt
}

func NewRotate(dx, dy, dz interface{}, children ...Stmt) *Rotate {
	return &Rotate{
		dx:       dx,
		dy:       dy,
		dz:       dz,
		children: children,
	}
}

func (r *Rotate) Add(s Stmt) *Rotate {
	r.children = append(r.children, s)
	return r
}

func (r *Rotate) EmitStmt(ctx context.Context, w io.Writer) error {
	indent := GetIndent(ctx)
	fmt.Fprintf(w, `%srotate([%#v, %#v, %#v])`, indent, r.dx, r.dy, r.dz)

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

func (l *LinearExtrude) EmitStmt(ctx context.Context, w io.Writer) error {
	ctx = context.WithValue(ctx, identAssignment{}, false)
	fmt.Fprintf(w, `%slinear_extrude(height=`, GetIndent(ctx))
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
		emitFn(ctx, w, l.fn)
	}
	fmt.Fprint(w, `)`)
	return emitChildren(ctx, w, l.children)
}
