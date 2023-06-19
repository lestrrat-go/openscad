package dsl_test

import (
	"os"
	"testing"

	"github.com/lestrrat-go/openscad/dsl"
	"gonum.org/v1/plot/tools/bezier"
	"gonum.org/v1/plot/vg"
)

func Example() {
	width := dsl.Variable("width").Value(30)
	stmts := dsl.Stmts(
		width,
		dsl.Module("foobar").
			Parameters(width).
			Actions(
				dsl.Rotate(
					dsl.List(0, 180, 0),
					dsl.Translate(
						dsl.List(10, 10, 10),
						dsl.Cube(width, 40, 5).Fn(24),
						dsl.Cube(5, 40, width),
						dsl.Cylinder(10, 5, 15).Fa(12),
					),
				),
			),
		dsl.Call("foobar"),
	)

	openscad.Emit(stmts, os.Stdout)
	//OUTPUT:
	// width=30;
	//
	// module foobar(width=30)
	// {
	//   rotate([0, 180, 0])
	//     translate([10, 10, 10])
	//     {
	//       cube([width, 40, 5], $fn=24);
	//       cube([5, 40, width]);
	//       cylinder(h=10, r1=5, r2=15, $fa=12);
	//     }
	// }
	// foobar();
}

func ExampleBezier2D() {
	crv := bezier.New(
		vg.Point{0, 0}, vg.Point{20, 2}, vg.Point{40, -1}, vg.Point{90, -3},
	)

	var pts openscad.Point2DList
	for _, pt := range crv.Curve(make([]vg.Point, 90/0.02)) {
		pts.Add(dsl.Point2D(pt.X, pt.Y))
	}

	points := dsl.Variable("points").Value(pts)
	stmts := dsl.Stmts(
		points,
		dsl.Polygon(points, nil),
	)

	openscad.Emit(stmts, os.Stdout)
}

func TestOperatorPrecedence(t *testing.T) {
	right := dsl.Add(3, 4)
	left := dsl.Mul(2, right)

	t.Logf("%t", left.BindPrecedence() > right.BindPrecedence())
	t.Logf("%#v", left.Rearrange(right))
}
