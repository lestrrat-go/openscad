package dsl

import (
	"github.com/lestrrat-go/openscad/ast"
)

func Lookup(key, values interface{}) *ast.LookupStmt {
	return ast.NewLookup(key, values)
}

func Call(name string, parameters ...interface{}) *ast.Call {
	call := ast.NewCall(name)
	if len(parameters) > 0 {
		call.Parameters(parameters...)
	}
	return call
}

func Children() *ast.Children {
	return ast.NewChildren()
}

func Concat(values ...interface{}) *ast.Call {
	return ast.NewCall("concat").Parameters(values...)
}

func Declare(name string, value interface{}) *ast.Declare {
	return ast.NewDeclare(Variable(name).Value(value))
}

func For(vars ...*ast.LoopVar) *ast.ForBlock {
	return ast.NewFor(vars)
}

func ForExpr(vars ...*ast.LoopVar) *ast.ForExpr {
	return ast.NewForExpr(vars)
}

func ForRange(start, end interface{}) *ast.ForRange {
	return ast.NewForRange(start, end)
}

func Function(name string) *ast.Function {
	return ast.NewFunction(name)
}

func Group(expr interface{}) *ast.Group {
	return ast.NewGroup(expr)
}

func Include(name string) *ast.Include {
	return ast.NewInclude(name)
}

func Index(v, index interface{}) *ast.Index {
	return ast.NewIndex(v, index)
}

func Len(v interface{}) *ast.Call {
	return ast.NewCall("len").Parameters(v)
}

func LetBlock(vars ...*ast.Variable) *ast.LetBlock {
	return ast.NewLetBlock(vars...)
}

func LetExpr(vars ...*ast.Variable) *ast.LetExpr {
	return ast.NewLetExpr(vars...)
}

func List(values ...interface{}) []interface{} {
	return values
}

func LoopVar(v *ast.Variable, expr interface{}) *ast.LoopVar {
	return ast.NewLoopVar(v, expr)
}

func Module(name string) *ast.Module {
	return ast.NewModule(name)
}

func Render() *ast.Render {
	return ast.NewRender()
}

func Stmts(stmts ...ast.Stmt) ast.Stmts {
	return ast.Stmts(stmts)
}

func Ternary(cond, left, right interface{}) *ast.TernaryOp {
	return ast.NewTernaryOp(cond, left, right)
}

func Use(name string) *ast.Use {
	return ast.NewUse(name)
}

func Variable(name string) *ast.Variable {
	return ast.NewVariable(name)
}
