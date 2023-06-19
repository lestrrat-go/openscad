package constants

import (
	"github.com/lestrrat-go/openscad/ast"
	"github.com/lestrrat-go/openscad/dsl"
)

var Cm = dsl.Variable("cm").Value(10)
var Inch = dsl.Variable("inch").Value(25.4)

func init() {
	ast.Register("constants.scad", Constants())
}

func Constants() ast.Stmt {
	return dsl.Stmts(
		Cm, Inch,
	)
}
