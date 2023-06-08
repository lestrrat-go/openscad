package dsl

import (
	"github.com/lestrrat-go/openscad"
)

func Polygon(pts, paths interface{}) *openscad.Polygon {
	return openscad.NewPolygon(pts, paths)
}

func Point2D(x, y interface{}) *openscad.Point2D {
	return openscad.NewPoint2D(x, y)
}

func Circle(radius interface{}) *openscad.Circle {
	return openscad.NewCircle(radius)
}

// Cube creates a cube with the given dimensions.
// There is no Cube(size=XXX), only Cube([X,Y,Z])
func Cube(x, y, z interface{}) *openscad.Cube {
	return openscad.NewCube(x, y, z)
}

func Cylinder(height, radius1, radius2 interface{}) *openscad.Cylinder {
	return openscad.NewCylinder(height, radius1, radius2)
}

func Sphere(radius interface{}) *openscad.Sphere {
	return openscad.NewSphere(radius)
}
