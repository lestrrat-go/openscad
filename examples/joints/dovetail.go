package joints

import (
	"embed"

	"github.com/lestrrat-go/openscad"
	"github.com/lestrrat-go/openscad/ast"
)

func init() {
	ast.Register("dovetail.scad", Dovetail())
}

//go:embed dovetail.scad
var dovetailSrc embed.FS

func Dovetail() ast.Stmt {
	stmts, err := openscad.ParseFile(`dovetail.scad`, openscad.WithFS(dovetailSrc))
	if err != nil {
		panic(err)
	}
	return stmts
}
