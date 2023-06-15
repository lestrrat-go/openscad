package openscad

import (
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

func (l Point2DList) EmitExpr(ctx *EmitContext, w io.Writer) error {
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

func (p *Point2D) EmitExpr(ctx *EmitContext, w io.Writer) error {
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

func (p *Polygon) EmitStmt(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, `%spolygon(points=`, ctx.Indent())
	ctx = ctx.WithAllowAssignment(false)
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

func emitCenter(w io.Writer, ptr *bool) {
	if ptr != nil && *ptr {
		fmt.Fprintf(w, `, $fn=%t`, *ptr)
	}
}

func emitFs(w io.Writer, ptr *int) {
	emitInt(w, `$fs`, ptr)
}
func emitFa(w io.Writer, ptr *int) {
	emitInt(w, `$fa`, ptr)
}
func emitFn(w io.Writer, ptr *int) {
	emitInt(w, `$fn`, ptr)
}

func emitInt(w io.Writer, name string, ptr *int) {
	if ptr != nil {
		fmt.Fprintf(w, `, %s=%d`, name, *ptr)
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

func (c *Cube) EmitStmt(ctx *EmitContext, w io.Writer) error {
	ctx = ctx.WithAllowAssignment(false)
	fmt.Fprintf(w, `%scube([`, ctx.Indent())
	emitValue(ctx, w, c.width)
	fmt.Fprintf(w, `, `)
	emitValue(ctx, w, c.depth)
	fmt.Fprintf(w, `, `)
	emitValue(ctx, w, c.height)
	fmt.Fprintf(w, `]`)

	emitCenter(w, c.center)
	emitFn(w, c.fn)
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

func (c *Cylinder) EmitStmt(ctx *EmitContext, w io.Writer) error {
	ctx = ctx.WithAllowAssignment(false)
	fmt.Fprintf(w, `%scylinder(h=`, ctx.Indent())
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
	emitCenter(w, c.center)
	emitFa(w, c.fa)
	emitFs(w, c.fs)
	emitFn(w, c.fn)
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

func (c *Children) EmitStmt(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, `%schildren(`, ctx.Indent())
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

func (s *Sphere) EmitStmt(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, `%ssphere(r=`, ctx.Indent())
	if s.radius == nil {
		return fmt.Errorf("radius must be specified")
	}
	emitValue(ctx, w, s.radius)
	emitFa(w, s.fa)
	emitFs(w, s.fs)
	emitFn(w, s.fn)
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

func (c *Circle) EmitExpr(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, `circle(r=`)
	if c.radius == nil {
		return fmt.Errorf("radius must be specified")
	}
	if err := emitExpr(ctx.WithAllowAssignment(false), w, c.radius); err != nil {
		return err
	}
	emitFa(w, c.fa)
	emitFn(w, c.fn)
	emitFs(w, c.fs)
	fmt.Fprintf(w, `)`)
	return nil
}

func (c *Circle) EmitStmt(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, `%s`, ctx.Indent())
	if err := c.EmitExpr(ctx, w); err != nil {
		return err
	}
	fmt.Fprintf(w, `;`)
	return nil
}

type Polyhedron struct {
	points    interface{}
	faces     interface{}
	convexity interface{}
}

func NewPolyhedron(points, faces interface{}) *Polyhedron {
	return &Polyhedron{
		points: points,
		faces:  faces,
	}
}

func (p *Polyhedron) Convexity(v interface{}) *Polyhedron {
	p.convexity = v
	return p
}

func (p *Polyhedron) EmitStmt(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, `%spolyhedron(points=`, ctx.Indent())
	ctx = ctx.WithAllowAssignment(false)
	if p.points == nil {
		return fmt.Errorf("points must be specified")
	}
	if err := emitExpr(ctx, w, p.points); err != nil {
		return err
	}
	fmt.Fprintf(w, `, faces=`)
	if p.faces == nil {
		return fmt.Errorf("faces must be specified")
	}
	if err := emitExpr(ctx, w, p.faces); err != nil {
		return err
	}
	if p.convexity != nil {
		fmt.Fprintf(w, `, convexity=`)
		if err := emitExpr(ctx, w, p.convexity); err != nil {
			return err
		}
	}
	fmt.Fprintf(w, `);`)
	return nil
}
