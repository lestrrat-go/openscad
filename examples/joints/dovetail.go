package joints

import (
	"github.com/lestrrat-go/openscad"
	"github.com/lestrrat-go/openscad/dsl"
)

func init() {
	openscad.Register("dovetail.scad", Dovetail())
}

func Dovetail() openscad.Stmt {
	return dsl.Stmts(
		DovetailTopOffset(),
		DovetailTenon(),
	)
}

// DovetailTopOffset calculates the offset in the horizontal
// direction for the top of the dovetail. Assuming that
// the origin of the bottom of the dovetail is at (0,0),
// the top should be at (-offset, height)
func DovetailTopOffset() openscad.Stmt {
	base := dsl.Variable("base")
	height := dsl.Variable("height")
	bh := dsl.Variable("backtrack_height")

	return dsl.Stmts(
		dsl.Function("dovetail_backtrack_height").
			Parameters(base).
			Body(
				dsl.Mul(base, dsl.Sqrt(2)),
			),
		dsl.Function("dovetail_top_width").
			Parameters(base, height, bh).
			Body(
				dsl.Mul(base, dsl.Div(dsl.Add(bh, height), bh)),
			),
		dsl.Function("dovetail_top_offset").
			Parameters(base, height).
			Body(
				dsl.Div(
					dsl.Sub(
						dsl.Call("dovetail_top_width", base, height, dsl.Call("dovetail_backtrack_height", base)),
						base,
					),
					2,
				),
			),
	)
}

func DovetailTenon() openscad.Stmt {
	// calculate starting from the "bottom" (base)of the ari
	//
	//   top
	//  ____
	//  \__/
	//   bottom
	//
	// we first find the point perpendicular from the middle
	// of the bottom line, with length bottom*sqrt(2)
	base := dsl.Variable("base")
	height := dsl.Variable("height")
	depth := dsl.Variable("depth")
	xoffset := dsl.Variable("xoffset").Value(dsl.Call("dovetail_top_offset", base, height))
	return dsl.Module("dovetail_tenon").
		Parameters(base, height, depth).
		Actions(
			xoffset,
			dsl.LinearExtrude(depth, nil, nil, nil, nil).
				Add(dsl.Polygon(
					dsl.List(
						dsl.List(xoffset, 0),
						dsl.List(dsl.Add(xoffset, base), 0),
						dsl.List(dsl.Add(dsl.Mul(xoffset, 2), base), height),
						dsl.List(0, height),
					),
					nil,
				)),
		)
}
