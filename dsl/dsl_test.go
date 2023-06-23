package dsl_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/lestrrat-go/openscad/ast"
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

	if err := ast.Emit(stmts, os.Stdout); err != nil {
		fmt.Printf("failed to emit: %s\n", err)
	}
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

//nolint:govet
func ExampleBezier2D() {
	crv := bezier.New(
		vg.Point{X: 0, Y: 0}, vg.Point{X: 20, Y: 2}, vg.Point{X: 40, Y: -1}, vg.Point{X: 90, Y: -3},
	)

	var pts ast.Point2DList
	for _, pt := range crv.Curve(make([]vg.Point, 90/0.02)) {
		pts.Add(dsl.Point2D(pt.X, pt.Y))
	}

	points := dsl.Variable("points").Value(pts)
	stmts := dsl.Stmts(
		points,
		dsl.Polygon(points, nil),
	)

	if err := ast.Emit(stmts, os.Stdout); err != nil {
		fmt.Printf("failed to emit: %s\n", err)
	}
}

func TestOperatorPrecedence(t *testing.T) {
	right := dsl.Add(3, 4)
	left := dsl.Mul(2, right)

	t.Logf("%t", left.BindPrecedence() > right.BindPrecedence())
	t.Logf("%#v", left.Rearrange(right))
}
