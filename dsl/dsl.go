package dsl

import "github.com/lestrrat-go/openscad"

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

func Div(left, right interface{}) *openscad.BinaryOp {
	return openscad.NewBinaryOp("/", left, right)
}

func For(vars ...*openscad.LoopVar) *openscad.For {
	return openscad.NewFor(vars)
}

func ForRange(start, end interface{}) *openscad.ForRange {
	return openscad.NewForRange(start, end)
}

func Function(name string) *openscad.Function {
	return openscad.NewFunction(name)
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

func Mul(left, right interface{}) *openscad.BinaryOp {
	return openscad.NewBinaryOp("*", left, right)
}

func Stmts(stmts ...openscad.Stmt) openscad.Stmts {
	return openscad.Stmts(stmts)
}

func Variable(name string) *openscad.Variable {
	return openscad.NewVariable(name)
}
