package dsl

import "github.com/lestrrat-go/openscad"

func Atan2(left, right interface{}) *openscad.Call {
	return openscad.NewCall("atan2").Parameters(left, right)
}

func Add(left, right interface{}) *openscad.BinaryOp {
	return openscad.NewBinaryOp("+", left, right)
}

func Ceil(v interface{}) *openscad.Call {
	return openscad.NewCall("ceil").Parameters(v)
}

func Cos(v interface{}) *openscad.Call {
	return openscad.NewCall("cos").Parameters(v)
}

func Div(left, right interface{}) *openscad.BinaryOp {
	return openscad.NewBinaryOp("/", left, right)
}

func EQ(left, right interface{}) *openscad.BinaryOp {
	return openscad.NewBinaryOp("==", left, right)
}

func Floor(v interface{}) *openscad.Call {
	return openscad.NewCall("floor").Parameters(v)
}

func GE(left, right interface{}) *openscad.BinaryOp {
	return openscad.NewBinaryOp(">=", left, right)
}

func GT(left, right interface{}) *openscad.BinaryOp {
	return openscad.NewBinaryOp(">", left, right)
}

func LE(left, right interface{}) *openscad.BinaryOp {
	return openscad.NewBinaryOp("<=", left, right)
}

func LT(left, right interface{}) *openscad.BinaryOp {
	return openscad.NewBinaryOp("<", left, right)
}

func Mod(left, right interface{}) *openscad.BinaryOp {
	return openscad.NewBinaryOp("%", left, right)
}

func Mul(left, right interface{}) *openscad.BinaryOp {
	return openscad.NewBinaryOp("*", left, right)
}

func Negative(v interface{}) *openscad.UnaryOp {
	return openscad.NewUnaryOp("-", v)
}

func PI() *openscad.Variable {
	return openscad.NewVariable("PI")
}

func Sin(v interface{}) *openscad.Call {
	return openscad.NewCall("sin").Parameters(v)
}

func Sqrt(v interface{}) *openscad.Call {
	return openscad.NewCall("sqrt").Parameters(v)
}

func Sub(left, right interface{}) *openscad.BinaryOp {
	return openscad.NewBinaryOp("-", left, right)
}

func Tan(v interface{}) *openscad.Call {
	return openscad.NewCall("tan").Parameters(v)
}
