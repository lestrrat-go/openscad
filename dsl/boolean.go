package dsl

import "github.com/lestrrat-go/openscad"

func Difference(stmts ...openscad.Stmt) *openscad.Difference {
	return openscad.NewDifference().Body(stmts...)
}

func Intersection(stmts ...openscad.Stmt) *openscad.Intersection {
	return openscad.NewIntersection().Body(stmts...)
}

func Union(stmts ...openscad.Stmt) *openscad.Union {
	return openscad.NewUnion().Body(stmts...)
}
