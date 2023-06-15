package threads

import (
	"github.com/lestrrat-go/openscad"
	"github.com/lestrrat-go/openscad/dsl"
)

func init() {
	openscad.Register("screwrod.scad", dsl.Stmts(
		Constants(),
		ThreadPitch(),
		ClosePoints(),
		ScrewThread(),
	))
}
