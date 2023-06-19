package openscad_test

import (
	"testing"

	"github.com/lestrrat-go/openscad"
	"github.com/lestrrat-go/openscad/ast"
	"github.com/lestrrat-go/openscad/dsl"
	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	testcases := []struct {
		Name     string
		Src      string
		Error    bool
		Expected interface{}
	}{
		{
			Name:     "include",
			Src:      "include <foo.scad>",
			Expected: dsl.Stmts(ast.NewInclude("foo.scad")),
		},
		{
			Name:     "assign array index lookup",
			Src:      "bar = foo[1];",
			Expected: dsl.Stmts(dsl.Variable("bar").Value(dsl.Index(dsl.Variable("foo"), 1.0))),
		},
		{
			Name:     "assing array index double lookup",
			Src:      "doubleLookupBar = doubleLookupFoo[1][2];",
			Expected: dsl.Stmts(dsl.Variable("doubleLookupBar").Value(dsl.Index(dsl.Index(dsl.Variable("doubleLookupFoo"), 1.0), 2.0))),
		},
		{
			Name:     "assign unary minus variable",
			Src:      "bar = -foo*3/2;",
			Expected: dsl.Stmts(dsl.Variable("bar").Value(dsl.Mul(dsl.Negative(dsl.Variable("foo")), dsl.Div(3.0, 2.0)))),
		},
		{
			Name:     "function declaration",
			Src:      "function double(x) = x * x + 2 * x - 1;",
			Expected: dsl.Stmts(dsl.Function("double").Parameters(dsl.Variable("x")).Body(dsl.Add(dsl.Mul(dsl.Variable("x"), dsl.Variable("x")), dsl.Sub(dsl.Mul(2.0, dsl.Variable("x")), 1.0)))),
		},
		{
			Name: "recursive function declaration",
			Src:  "function recurse_avg(arr, n=0, p=[0,0,0]) = (n>=len(arr)) ? p : recurse_avg(arr, n+1, p+(arr[n]-p)/(n+1));",
			//Src:      "function recurse_avg(arr, n=0, p=[0,0,0]) = (n>=len(arr)) ? p : recurse_avg(arr, n+1, p+(arr-p)/(n+1));",
			Expected: dsl.Stmts(dsl.Function("recurse_avg").
				Parameters(
					dsl.Variable("arr"),
					dsl.Variable("n").Value(0.0),
					dsl.Variable("p").Value(dsl.List(0.0, 0.0, 0.0)),
				).
				Body(
					dsl.Ternary(
						dsl.Group(dsl.GE(dsl.Variable("n"), dsl.Len(dsl.Variable("arr")))),
						dsl.Variable("p"),
						dsl.Call("recurse_avg",
							dsl.Variable("arr"),
							dsl.Add(dsl.Variable("n"), 1.0),
							dsl.Add(dsl.Variable("p"), dsl.Div(
								dsl.Group(dsl.Sub(dsl.Index(dsl.Variable("arr"), dsl.Variable("n")), dsl.Variable("p"))),
								dsl.Group(dsl.Add(dsl.Variable("n"), 1.0))),
							),
						),
					),
				)),
		},
		{
			Name: "multiple for ranges",
			Src:  `faces_loop = [ for (j=[0:N-2], i=[0:P-1], t=[0:1]) [loop_offset, loop_offset, loop_offset] + (t==0 ?  [j*P+i, (j+1)*P+i, (j+1)*P+(i+1)%P] : [j*P+i, (j+1)*P+(i+1)%P, j*P+(i+1)%P]) ];`,
			Expected: dsl.Stmts(
				dsl.Variable("faces_loop").Value(
					dsl.List(
						dsl.ForExpr(
							dsl.LoopVar(dsl.Variable("j"),
								dsl.ForRange(0.0, dsl.Sub(dsl.Variable("N"), 2.0))),
							dsl.LoopVar(dsl.Variable("i"),
								dsl.ForRange(0.0, dsl.Sub(dsl.Variable("P"), 1.0))),
							dsl.LoopVar(dsl.Variable("t"),
								dsl.ForRange(0.0, 1.0)),
						).Body(
							dsl.Add(
								dsl.List(dsl.Variable("loop_offset"), dsl.Variable("loop_offset"), dsl.Variable("loop_offset")),
								dsl.Group(
									dsl.Ternary(
										// t == 0
										dsl.EQ(dsl.Variable("t"), 0.0),
										dsl.List(
											// j*P+i
											dsl.Add(dsl.Mul(dsl.Variable("j"), dsl.Variable("P")), dsl.Variable("i")),
											// (j+1)*P+i
											dsl.Add(
												dsl.Mul(
													dsl.Group(
														dsl.Add(dsl.Variable("j"), 1.0),
													),
													dsl.Variable("P"),
												),
												dsl.Variable("i"),
											),
											dsl.Add(
												dsl.Mul(
													dsl.Group(
														dsl.Add(dsl.Variable("j"), 1.0),
													),
													dsl.Variable("P"),
												),
												dsl.Mod(
													dsl.Group(
														dsl.Add(dsl.Variable("i"), 1.0),
													),
													dsl.Variable("P"),
												),
											),
										),
										dsl.List(
											dsl.Add(dsl.Mul(dsl.Variable("j"), dsl.Variable("P")), dsl.Variable("i")),
											dsl.Add(
												dsl.Mul(
													dsl.Group(
														dsl.Add(dsl.Variable("j"), 1.0),
													),
													dsl.Variable("P"),
												),
												dsl.Mod(
													dsl.Group(
														dsl.Add(dsl.Variable("i"), 1.0),
													),
													dsl.Variable("P"),
												),
											),
											// j*P+(i+1)%P])
											dsl.Add(
												dsl.Mul(
													dsl.Variable("j"),
													dsl.Variable("P"),
												),
												dsl.Mod(
													dsl.Group(
														dsl.Add(dsl.Variable("i"), 1.0),
													),
													dsl.Variable("P"),
												),
											),
										),
									),
								),
							),
						),
					),
				),
			),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			v, err := openscad.Parse([]byte(tc.Src))

			if tc.Error {
				require.Error(t, err, "Parse should fail")
			} else {
				require.NoError(t, err, "Parse should succeed")
				require.Equal(t, tc.Expected, v, "Parse result should match")
			}
		})
	}
	/*
			const src = `
		// Test code
		include <foo.scad>
		cm = 10;
		inch = 2.54 * cm; // one inch is 2.54 cm
		global_var = 1;

		function double(x) = x * x + 2 * x - 1;
		function mod(x, y) = x % y == 1;

		module foo(a, b, c=0) {
			bar=1;
			baz="hello";
		}

		foo(1, 2);

		render()
			foo(2, 3, 4);

		translate([1, 2, 3]) {
			cube([10, 10, 10]);
			cylinder(r=5, h=10);
		}

		  points = [
		    for (i=[-1:NP])
		      (i<0) ? midbot :
		      ((i==NP) ? midtop :
		      pointarrays[floor(i/P)][i%P])
		  ];

		  for (yloc=[-foo*2/3,foo*2/3]) {
			translate([0, yloc+leeway, 0.25*cm])
				wood_piece(ring_width, leeway=[0, leeway,leeway]);
		}


		`
			stmts, err := openscad.Parse([]byte(src))
			log.Printf("%s", err)
			log.Printf("%#v", stmts)
			ast.Emit(stmts, os.Stdout)
	*/
}
