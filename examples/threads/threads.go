// Based on code from https://github.com/rcolyer/threads-scad

package threads

import (
	"github.com/lestrrat-go/openscad"
	"github.com/lestrrat-go/openscad/dsl"
)

var screwResolution = dsl.Variable("screw_resolution").Value(0.2)

func Constants() openscad.Stmt {
	return dsl.Stmts(
		screwResolution,
	)
}

func ThreadPitch() openscad.Stmt {
	diameter := dsl.Variable("diameter")
	return dsl.Function("ThreadPitch").
		Parameters(diameter).
		Body(
			dsl.Ternary(
				dsl.GE(diameter, 64),
				dsl.Lookup(diameter, dsl.List(
					dsl.List(2, 0.4),
					dsl.List(2.5, 0.45),
					dsl.List(3, 0.5),
					dsl.List(4, 0.7),
					dsl.List(5, 0.8),
					dsl.List(6, 1.0),
					dsl.List(7, 1.0),
					dsl.List(8, 1.25),
					dsl.List(10, 1.5),
					dsl.List(12, 1.75),
					dsl.List(14, 2.0),
					dsl.List(16, 2.0),
					dsl.List(18, 2.5),
					dsl.List(20, 2.5),
					dsl.List(22, 2.5),
					dsl.List(24, 3.0),
					dsl.List(27, 3.0),
					dsl.List(30, 3.5),
					dsl.List(33, 3.5),
					dsl.List(36, 4.0),
					dsl.List(39, 4.0),
					dsl.List(42, 4.5),
					dsl.List(48, 5.0),
					dsl.List(52, 5.0),
					dsl.List(56, 5.5),
					dsl.List(60, 5.5),
					dsl.List(64, 6.0),
				)),
				dsl.Div(dsl.Mul(diameter, 6.0), 64),
			),
		)
}

func RecursiveAvg() openscad.Stmt {
	arr := dsl.Variable("arr")
	n := dsl.Variable("n").Value(0)
	p := dsl.Variable("p").Value(dsl.List(0, 0, 0))
	return dsl.Function("recurse_avg").
		Parameters(arr, n, p).
		Body(
			dsl.Ternary(
				dsl.GE(n, dsl.Len(arr)),
				p,
				dsl.Call("recurse_avg", arr, dsl.Add(n, 1), dsl.Add(p, dsl.Div(dsl.Sub(dsl.Index(arr, n), p), dsl.Add(n, 1)))),
			),
		)
}

func ClosePoints() openscad.Stmt {
	pointarrays := dsl.Variable("pointarrays")
	N := dsl.Variable("N").Value(dsl.Len(pointarrays))
	P := dsl.Variable("P").Value(dsl.Len(dsl.Index(pointarrays, 0)))
	NP := dsl.Variable("NP").Value(dsl.Mul(N, P))
	lastarr := dsl.Variable("lastarr").Value(dsl.Index(pointarrays, dsl.Sub(N, 1)))
	midbot := dsl.Variable("midbot").Value(dsl.Call("recurse_avg", dsl.Index(pointarrays, 0)))
	midtop := dsl.Variable("midtop").Value(dsl.Call("recurse_avg", dsl.Index(pointarrays, dsl.Sub(N, 1))))

	i := dsl.Variable("i")
	facesBot := dsl.Variable("faces_bot").Value(dsl.List(
		dsl.ForExpr(dsl.LoopVar(i, dsl.ForRange(0, dsl.Sub(P, 1)))).
			Body(
				dsl.List(0, dsl.Add(i, 1), dsl.Add(1, dsl.Mod(dsl.Add(i, 1), dsl.Len(dsl.Index(pointarrays, 0))))),
			),
	))

	loopOffset := dsl.Variable("loop_offset").Value(1)
	botLen := dsl.Variable("bot_len").Value(dsl.Add(loopOffset, P))
	j := dsl.Variable("j")
	t := dsl.Variable("t")
	facesLoop := dsl.Variable("faces_loop").Value(dsl.List(
		dsl.ForExpr(
			dsl.LoopVar(j, dsl.ForRange(0, dsl.Sub(N, 2))),
			dsl.LoopVar(i, dsl.ForRange(0, dsl.Sub(P, 1))),
			dsl.LoopVar(t, dsl.ForRange(0, 1)),
		).Body(
			dsl.Add(
				dsl.List(loopOffset, loopOffset, loopOffset),
				dsl.Ternary(
					dsl.EQ(t, 0),
					dsl.List(
						dsl.Add(dsl.Mul(j, P), i), dsl.Add(dsl.Mul(dsl.Add(j, 1), P), 1), dsl.Add(dsl.Mul(dsl.Add(j, 1), P), dsl.Mod(dsl.Add(i, 1), P)),
					),
					dsl.List(
						dsl.Add(dsl.Mul(j, P), i), dsl.Add(dsl.Mul(dsl.Add(j, 1), P), dsl.Mod(dsl.Add(i, 1), P)), dsl.Add(dsl.Mul(j, P), dsl.Mod(dsl.Add(i, 1), P)),
					),
				),
			),
		),
	))

	topOffset := dsl.Variable("top_offset").Value(
		dsl.Sub(dsl.Add(loopOffset, NP), P),
	)
	midtopOffset := dsl.Variable("midtop_offset").Value(dsl.Add(topOffset, P))

	facesTop := dsl.Variable("faces_top").Value(dsl.List(
		dsl.ForExpr(dsl.LoopVar(i, dsl.ForRange(0, dsl.Sub(P, 1)))).
			Body(
				dsl.List(midtopOffset, dsl.Add(topOffset, dsl.Mod(dsl.Add(i, 1), P)), dsl.Add(topOffset, i)),
			),
	))

	points := dsl.Variable("points").Value(dsl.List(
		dsl.ForExpr(dsl.LoopVar(i, dsl.ForRange(-1, NP))).
			Body(
				dsl.Ternary(
					dsl.LE(i, 0),
					midbot,
					dsl.Ternary(
						dsl.EQ(i, NP),
						midtop,
						dsl.Index(dsl.Index(pointarrays, dsl.Floor(dsl.Div(i, P))), dsl.Mod(i, P)),
					),
				),
			),
	))
	faces := dsl.Variable("faces").Value(
		dsl.Concat(facesBot, facesLoop, facesTop),
	)

	return dsl.Module("ClosePoints").
		Parameters(pointarrays).
		Body(
			RecursiveAvg(),
			N,
			P,
			NP,
			lastarr,
			midbot,
			midtop,
			facesBot,
			loopOffset,
			botLen,
			facesLoop,
			topOffset,
			midtopOffset,
			facesTop,
			points,
			faces,
			dsl.Polyhedron(points, faces),
		)
}

func ScrewThread() openscad.Stmt {
	outerDiam := dsl.Variable("outer_diam")
	height := dsl.Variable("height")
	pitch := dsl.Variable("pitch").Value(0)
	toothAngle := dsl.Variable("tooth_angle").Value(30)
	tolerance := dsl.Variable("tolerance").Value(0.4)
	tipHeight := dsl.Variable("tip_height").Value(0)
	toothHeight := dsl.Variable("tooth_height").Value(0)
	tipMinFract := dsl.Variable("tip_min_fract").Value(0)

	outerDiamCor := dsl.Variable("outer_diam_cor").Value(dsl.Add(outerDiam, dsl.Mul(0.25, tolerance)))
	innerDiam := dsl.Variable("inner_diam").Value(dsl.Sub(outerDiam, dsl.Div(toothHeight, dsl.Tan(toothAngle))))
	or := dsl.Variable("or").Value(
		dsl.Ternary(
			dsl.LT(outerDiamCor, screwResolution),
			dsl.Div(screwResolution, 2),
			dsl.Div(outerDiamCor, 2),
		),
	)
	ir := dsl.Variable("ir").Value(
		dsl.Ternary(
			dsl.LT(innerDiam, screwResolution),
			dsl.Div(screwResolution, 2),
			dsl.Div(innerDiam, 2),
		),
	)

	stepsPerLoopTry := dsl.Variable("steps_per_loop_try").Value(
		dsl.Ceil(dsl.Div(dsl.Mul(2, dsl.Mul(dsl.PI(), or)), screwResolution)),
	)
	stepsPerLoop := dsl.Variable("steps_per_loop").Value(
		dsl.Ternary(
			dsl.LT(stepsPerLoopTry, 4),
			4,
			stepsPerLoopTry,
		),
	)
	hsExt := dsl.Variable("hs_ext").Value(2)
	hsteps := dsl.Variable("hsteps").Value(
		dsl.Ceil(dsl.Add(dsl.Div(dsl.Mul(3, height), pitch), dsl.Mul(2, hsExt))),
	)
	extent := dsl.Variable("extent").Value(dsl.Sub(or, ir))

	tipStart := dsl.Variable("tip_start").Value(
		dsl.Sub(height, tipHeight))
	tipHeightSc := dsl.Variable("tip_height_sc").Value(
		dsl.Div(tipHeight, dsl.Sub(1, tipMinFract)),
	)
	tipHeightIr := dsl.Variable("tip_height_ir").Value(
		dsl.Ternary(
			dsl.GT(tipHeightSc, dsl.Div(toothHeight, 2)),
			dsl.Sub(tipHeightSc, dsl.Div(toothHeight, 2)),
			tipHeightSc,
		),
	)
	tipHeightW := dsl.Variable("tip_height_w").Value(
		dsl.Ternary(
			dsl.GT(tipHeightSc, toothHeight),
			toothHeight,
			tipHeightSc,
		),
	)
	tipWstart := dsl.Sub(dsl.Sub(dsl.Add(height, tipHeightSc), tipHeight), tipHeightW)

	return dsl.Module("ScrewThread").
		Parameters(outerDiam, height, pitch, toothAngle, tolerance, tipHeight, toothHeight, tipMinFract).
		Body(
			dsl.Variable("pitch").Value(
				dsl.Ternary(
					dsl.EQ(pitch, 0),
					dsl.Call("ThreadPitch", outerDiam),
					pitch,
				),
			),
			dsl.Variable("tooth_angle").Value(
				dsl.Ternary(
					dsl.EQ(toothHeight, 0),
					pitch,
					toothHeight,
				),
			),
			dsl.Variable("tip_min_fract").Value(
				dsl.Ternary(
					dsl.LT(tipMinFract, 0),
					0,
					dsl.Ternary(
						dsl.GT(tipMinFract, 0.9999),
						0.9999,
						tipMinFract,
					),
				),
			),
			outerDiamCor,
			innerDiam,
			or,
			ir,
			dsl.Variable("height").Value(
				dsl.Ternary(
					dsl.LT(height, screwResolution),
					screwResolution,
					height,
				),
			),
			stepsPerLoopTry,
			stepsPerLoop,
			hsExt,
			hsteps,
			extent,
			tipStart,
			tipHeightSc,
			tipHeightIr,
			tipHeightW,
			tipWstart,
			toothWidthFunc(),
		)
}

func toothWidthFunc() openscad.Stmt {
	angle := dsl.Variable("angle")
	height := dsl.Variable("height")
	pitch := dsl.Variable("pitch")
	toothHeight := dsl.Variable("tooth_height")
	extent := dsl.Variable("extent")

	angFull := dsl.Variable("ang_full").Value(
		dsl.Div(dsl.Mul(height, 360), angle),
	)
	angPn := dsl.Variable("angPn").Value(dsl.Atan2(dsl.Sin(angFull), dsl.Cos(angFull)))

	return dsl.Function("tooth_width").
		Parameters(
			angle,
			height,
			pitch,
			toothHeight,
			dsl.Variable("extent"),
		).Body(
		dsl.Let(
			angFull,
			angPn,
		).Body(
			extent,
		),
	)
}
