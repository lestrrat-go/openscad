package openscad

import (
	"context"
	"fmt"
	"io"
)

type Point2D struct {
	x, y interface{}
}

type Point2DList []*Point2D

func (l *Point2DList) Add(pt *Point2D) {
	*l = append(*l, pt)
}

func (l Point2DList) EmitExpr(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, `[`)
	for i, pt := range l {
		if i > 0 {
			fmt.Fprintf(w, `, `)
		}
		if err := pt.EmitExpr(ctx, w); err != nil {
			return err
		}
	}
	fmt.Fprintf(w, `]`)
	return nil
}

func NewPoint2D(x, y interface{}) *Point2D {
	return &Point2D{
		x: x,
		y: y,
	}
}

func (p *Point2D) EmitExpr(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, `[%#v, %#v]`, p.x, p.y)
	return nil
}

type Polygon struct {
	points interface{}
	paths  interface{}
}

func NewPolygon(points, paths interface{}) *Polygon {
	return &Polygon{
		points: points,
		paths:  paths,
	}
}

func (p *Polygon) EmitStmt(ctx context.Context, w io.Writer) error {
	indent := GetIndent(ctx)
	fmt.Fprintf(w, `%spolygon(points=`, indent)
	ctx = context.WithValue(ctx, identAssignment{}, false)
	if err := emitValue(ctx, w, p.points); err != nil {
		return err
	}
	if p.paths != nil {
		fmt.Fprintf(w, `, paths=`)
		if err := emitValue(ctx, w, p.paths); err != nil {
			return err
		}
	}
	fmt.Fprint(w, `);`)
	return nil
}

func emitCenter(ctx context.Context, w io.Writer, ptr *bool) {
	if ptr != nil && *ptr {
		fmt.Fprintf(w, `, $fn=%t`, *ptr)
	}
}

func emitFs(ctx context.Context, w io.Writer, ptr *int) {
	emitInt(ctx, w, identFs{}, `$fs`, ptr)
}
func emitFa(ctx context.Context, w io.Writer, ptr *int) {
	emitInt(ctx, w, identFa{}, `$fa`, ptr)
}
func emitFn(ctx context.Context, w io.Writer, ptr *int) {
	emitInt(ctx, w, identFn{}, `$fn`, ptr)
}

func emitInt(ctx context.Context, w io.Writer, ident interface{}, name string, ptr *int) {
	if ptr != nil {
		fmt.Fprintf(w, `, %s=%d`, name, *ptr)
	} else {
		var v int
		if GetValue(ctx, ident, &v) == nil {
			fmt.Fprintf(w, `, %s=%d`, name, v)
		}
	}
}

type Cube struct {
	width, depth, height interface{}
	center               *bool
	fn                   *int
}

func NewCube(width, depth, height interface{}) *Cube {
	return &Cube{
		width:  width,
		depth:  depth,
		height: height,
	}
}

func (c *Cube) Center(v bool) *Cube {
	c.center = &v
	return c
}

func (c *Cube) Fn(v int) *Cube {
	c.fn = &v
	return c
}

func (c *Cube) EmitStmt(ctx context.Context, w io.Writer) error {
	ctx = context.WithValue(ctx, identAssignment{}, false)
	fmt.Fprintf(w, `%scube([`, GetIndent(ctx))
	emitValue(ctx, w, c.width)
	fmt.Fprintf(w, `, `)
	emitValue(ctx, w, c.depth)
	fmt.Fprintf(w, `, `)
	emitValue(ctx, w, c.height)
	fmt.Fprintf(w, `]`)

	emitCenter(ctx, w, c.center)
	emitFn(ctx, w, c.fn)
	fmt.Fprintf(w, `);`) // cubes are always terminated with a semicolon
	return nil
}

type Cylinder struct {
	height, radius1, radius2 interface{}
	center                   *bool
	fa                       *int
	fs                       *int
	fn                       *int
}

// To omit radius2, pass nil
func NewCylinder(height, radius1, radius2 interface{}) *Cylinder {
	return &Cylinder{
		height:  height,
		radius1: radius1,
		radius2: radius2,
	}
}

func (c *Cylinder) Center(v bool) *Cylinder {
	c.center = &v
	return c
}

func (c *Cylinder) Fa(v int) *Cylinder {
	c.fa = &v
	return c
}

func (c *Cylinder) Fs(v int) *Cylinder {
	c.fs = &v
	return c
}

func (c *Cylinder) Fn(v int) *Cylinder {
	c.fn = &v
	return c
}

func (c *Cylinder) EmitStmt(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, `%scylinder(h=`, GetIndent(ctx))
	if c.height == nil {
		return fmt.Errorf("height must be specified")
	}
	emitValue(ctx, w, c.height)

	if c.radius1 == nil {
		return fmt.Errorf("radius1 must be specified")
	}

	if c.radius2 == nil {
		fmt.Fprint(w, `, r=`)
		emitValue(ctx, w, c.radius1)
	} else {
		fmt.Fprint(w, `, r1=`)
		emitValue(ctx, w, c.radius1)
		fmt.Fprint(w, `, r2=`)
		emitValue(ctx, w, c.radius2)
	}
	emitCenter(ctx, w, c.center)
	emitFa(ctx, w, c.fa)
	emitFs(ctx, w, c.fs)
	emitFn(ctx, w, c.fn)
	fmt.Fprintf(w, `);`) // cylinders are always terminated with a semicolon
	return nil
}

// Creates a call to the children() module.
type Children struct {
	idx *int
}

func NewChildren() *Children {
	return &Children{}
}

func (c *Children) Index(idx int) *Children {
	c.idx = &idx
	return c
}

func (c *Children) EmitStmt(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, `%schildren(`, GetIndent(ctx))
	if c.idx != nil {
		fmt.Fprintf(w, `%d`, *c.idx)
	}
	fmt.Fprintf(w, `);`)
	return nil
}

type Sphere struct {
	radius interface{}
	fa     *int
	fs     *int
	fn     *int
}

func NewSphere(radius interface{}) *Sphere {
	return &Sphere{
		radius: radius,
	}
}

func (s *Sphere) Fa(v int) *Sphere {
	s.fa = &v
	return s
}

func (s *Sphere) Fs(v int) *Sphere {
	s.fs = &v
	return s
}

func (s *Sphere) Fn(v int) *Sphere {
	s.fn = &v
	return s
}

func (s *Sphere) EmitStmt(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, `%ssphere(r=`, GetIndent(ctx))
	if s.radius == nil {
		return fmt.Errorf("radius must be specified")
	}
	emitValue(ctx, w, s.radius)
	emitFa(ctx, w, s.fa)
	emitFs(ctx, w, s.fs)
	emitFn(ctx, w, s.fn)
	fmt.Fprintf(w, `);`)
	return nil
}

type Circle struct {
	radius interface{}
	fa     *int
	fn     *int
	fs     *int
}

func NewCircle(radius interface{}) *Circle {
	return &Circle{
		radius: radius,
	}
}

func (c *Circle) Fa(v int) *Circle {
	c.fa = &v
	return c
}

func (c *Circle) Fn(v int) *Circle {
	c.fn = &v
	return c
}

func (c *Circle) Fs(v int) *Circle {
	c.fs = &v
	return c
}

func (c *Circle) EmitExpr(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, `circle(r=`)
	if c.radius == nil {
		return fmt.Errorf("radius must be specified")
	}
	emitValue(ctx, w, c.radius)
	emitFa(ctx, w, c.fa)
	emitFn(ctx, w, c.fn)
	emitFs(ctx, w, c.fs)
	fmt.Fprintf(w, `)`)
	return nil
}

func (c *Circle) EmitStmt(ctx context.Context, w io.Writer) error {
	fmt.Fprintf(w, `%s`, GetIndent(ctx))
	if err := c.EmitExpr(ctx, w); err != nil {
		return err
	}
	fmt.Fprintf(w, `;`)
	return nil
}
