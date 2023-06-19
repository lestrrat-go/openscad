package dsl

import "github.com/lestrrat-go/openscad/ast"

func Hull(stmts ...ast.Stmt) *ast.Hull {
	return ast.NewHull().Body(stmts...)
}

func LinearExtrude(height, center, convexity, twist, slices interface{}) *ast.LinearExtrude {
	return ast.NewLinearExtrude(height, center, convexity, twist, slices)
}

func Translate(v interface{}, children ...ast.Stmt) *ast.Translate {
	return ast.NewTranslate(v, children...)
}

func Rotate(v interface{}, children ...ast.Stmt) *ast.Rotate {
	return ast.NewRotate(v, children...)
}
