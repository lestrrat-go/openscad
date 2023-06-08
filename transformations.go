package openscad

import (
	"context"
	"fmt"
	"io"
)

type Translate struct {
	x, y, z  interface{}
	children []Stmt
}

func NewTranslate(x, y, z interface{}, children ...Stmt) *Translate {
	return &Translate{
		x:        x,
		y:        y,
		z:        z,
		children: children,
	}
}

func (t *Translate) Add(s Stmt) *Translate {
	t.children = append(t.children, s)
	return t
}

func (t *Translate) EmitStmt(ctx context.Context, w io.Writer) error {
	indent := GetIndent(ctx)
	fmt.Fprintf(w, `%stranslate([%#v, %#v, %#v])`, indent, t.x, t.y, t.z)
	return emitChildren(ctx, w, t.children)
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
