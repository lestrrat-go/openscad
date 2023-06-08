package dsl

import "github.com/lestrrat-go/openscad"

func Union(stmts ...openscad.Stmt) *openscad.Union {
	return openscad.NewUnion(stmts...)
}
