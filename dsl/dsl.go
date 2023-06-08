package dsl

import "github.com/lestrrat-go/openscad"

func Stmts(stmts ...openscad.Stmt) openscad.Stmts {
	return openscad.Stmts(stmts)
}

func Module(name string) *openscad.Module {
	return openscad.NewModule(name)
}

func Variable(name string) *openscad.Variable {
	return openscad.NewVariable(name)
}

func Call(name string, parameters ...*openscad.Variable) *openscad.Call {
	call := openscad.NewCall(name)
	if len(parameters) > 0 {
		call.Parameters(parameters...)
	}
	return call
}
