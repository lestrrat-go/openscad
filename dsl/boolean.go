package dsl

import "github.com/lestrrat-go/openscad/ast"

func Difference(stmts ...ast.Stmt) *ast.Difference {
	return ast.NewDifference().Body(stmts...)
}

func Intersection(stmts ...ast.Stmt) *ast.Intersection {
	return ast.NewIntersection().Body(stmts...)
}

func Union(stmts ...ast.Stmt) *ast.Union {
	return ast.NewUnion().Body(stmts...)
}
