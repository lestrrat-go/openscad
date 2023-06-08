package dsl

import "github.com/lestrrat-go/openscad"

func Cos(v interface{}) *openscad.Call {
	return openscad.NewCall("cos").Parameters(v)
}

func Sin(v interface{}) *openscad.Call {
	return openscad.NewCall("sin").Parameters(v)
}
