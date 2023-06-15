package dsl

import (
	"github.com/lestrrat-go/openscad"
)

func Lookup(key, values interface{}) *openscad.LookupStmt {
	return openscad.NewLookup(key, values)
}

func Call(name string, parameters ...interface{}) *openscad.Call {
	call := openscad.NewCall(name)
	if len(parameters) > 0 {
		call.Parameters(parameters...)
	}
	return call
}

func Children() *openscad.Children {
	return openscad.NewChildren()
}

func Concat(values ...interface{}) *openscad.Call {
	return openscad.NewCall("concat").Parameters(values...)
}

func Declare(name string, value interface{}) *openscad.Declare {
	return openscad.NewDeclare(Variable(name).Value(value))
}

func For(vars ...*openscad.LoopVar) *openscad.ForBlock {
	return openscad.NewFor(vars)
}

func ForExpr(vars ...*openscad.LoopVar) *openscad.ForExpr {
	return openscad.NewForExpr(vars)
}

func ForRange(start, end interface{}) *openscad.ForRange {
	return openscad.NewForRange(start, end)
}

func Function(name string) *openscad.Function {
	return openscad.NewFunction(name)
}

func Include(name string) *openscad.Include {
	return openscad.NewInclude(name)
}

func Index(v, index interface{}) *openscad.Index {
	return openscad.NewIndex(v, index)
}

func Len(v interface{}) *openscad.Len {
	return openscad.NewLen(v)
}

func Let(vars ...*openscad.Variable) *openscad.Let {
	return openscad.NewLet(vars...)
}

func List(values ...interface{}) []interface{} {
	return values
}

func LoopVar(v *openscad.Variable, expr interface{}) *openscad.LoopVar {
	return openscad.NewLoopVar(v, expr)
}

func Module(name string) *openscad.Module {
	return openscad.NewModule(name)
}

func Render() *openscad.Render {
	return openscad.NewRender()
}

func Stmts(stmts ...openscad.Stmt) openscad.Stmts {
	return openscad.Stmts(stmts)
}

func Ternary(cond, left, right interface{}) *openscad.TernaryOp {
	return openscad.NewTernaryOp(cond, left, right)
}

func Use(name string) *openscad.Use {
	return openscad.NewUse(name)
}

func Variable(name string) *openscad.Variable {
	return openscad.NewVariable(name)
}
