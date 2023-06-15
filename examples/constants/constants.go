package constants

import (
	"github.com/lestrrat-go/openscad"
	"github.com/lestrrat-go/openscad/dsl"
)

var Cm = dsl.Variable("cm").Value(10)
var Inch = dsl.Variable("inch").Value(25.4)

func init() {
	openscad.Register("constants.scad", Constants())
}

func Constants() openscad.Stmt {
	return dsl.Stmts(
		Cm, Inch,
	)
}
