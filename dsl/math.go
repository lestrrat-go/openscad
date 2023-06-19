package dsl

import ast "github.com/lestrrat-go/openscad/ast"

func Atan2(left, right interface{}) *ast.Call {
	return ast.NewCall("atan2").Parameters(left, right)
}

func Add(left, right interface{}) *ast.BinaryOp {
	return ast.NewBinaryOp("+", left, right)
}

func Ceil(v interface{}) *ast.Call {
	return ast.NewCall("ceil").Parameters(v)
}

func Cos(v interface{}) *ast.Call {
	return ast.NewCall("cos").Parameters(v)
}

func Div(left, right interface{}) *ast.BinaryOp {
	return ast.NewBinaryOp("/", left, right)
}

func EQ(left, right interface{}) *ast.BinaryOp {
	return ast.NewBinaryOp("==", left, right)
}

func Floor(v interface{}) *ast.Call {
	return ast.NewCall("floor").Parameters(v)
}

func GE(left, right interface{}) *ast.BinaryOp {
	return ast.NewBinaryOp(">=", left, right)
}

func GT(left, right interface{}) *ast.BinaryOp {
	return ast.NewBinaryOp(">", left, right)
}

func LE(left, right interface{}) *ast.BinaryOp {
	return ast.NewBinaryOp("<=", left, right)
}

func LT(left, right interface{}) *ast.BinaryOp {
	return ast.NewBinaryOp("<", left, right)
}

func Mod(left, right interface{}) *ast.BinaryOp {
	return ast.NewBinaryOp("%", left, right)
}

func Mul(left, right interface{}) *ast.BinaryOp {
	return ast.NewBinaryOp("*", left, right)
}

func Negative(v interface{}) *ast.UnaryOp {
	return ast.NewUnaryOp("-", v)
}

func PI() *ast.Variable {
	return ast.NewVariable("PI")
}

func Sin(v interface{}) *ast.Call {
	return ast.NewCall("sin").Parameters(v)
}

func Sqrt(v interface{}) *ast.Call {
	return ast.NewCall("sqrt").Parameters(v)
}

func Sub(left, right interface{}) *ast.BinaryOp {
	return ast.NewBinaryOp("-", left, right)
}

func Tan(v interface{}) *ast.Call {
	return ast.NewCall("tan").Parameters(v)
}
