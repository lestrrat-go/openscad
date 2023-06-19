package ast_test

import "github.com/lestrrat-go/openscad/ast"

var _ ast.EmitOption = ast.WithAmalgamation()
var _ ast.EmitFileOption = ast.WithAmalgamation()
var _ ast.WriteFileOption = ast.WithAmalgamation()
