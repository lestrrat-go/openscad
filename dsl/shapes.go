package dsl

import (
	"github.com/lestrrat-go/openscad/ast"
)

func Polygon(pts, paths interface{}) *ast.Polygon {
	return ast.NewPolygon(pts, paths)
}

func Point2D(x, y interface{}) *ast.Point2D {
	return ast.NewPoint2D(x, y)
}

func Circle(radius interface{}) *ast.Circle {
	return ast.NewCircle(radius)
}

// Cube creates a cube with the given dimensions.
// There is no Cube(size=XXX), only Cube([X,Y,Z])
func Cube(x, y, z interface{}) *ast.Cube {
	return ast.NewCube(x, y, z)
}

func Cylinder(height, radius1, radius2 interface{}) *ast.Cylinder {
	return ast.NewCylinder(height, radius1, radius2)
}

func Polyhedron(points, triangles interface{}) *ast.Polyhedron {
	return ast.NewPolyhedron(points, triangles)
}

func Sphere(radius interface{}) *ast.Sphere {
	return ast.NewSphere(radius)
}
