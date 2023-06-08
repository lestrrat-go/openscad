package dsl

import "github.com/lestrrat-go/openscad"

func LinearExtrude(height, center, convexity, twist, slices interface{}) *openscad.LinearExtrude {
	return openscad.NewLinearExtrude(height, center, convexity, twist, slices)
}

func Translate(v interface{}, children ...openscad.Stmt) *openscad.Translate {
	return openscad.NewTranslate(v, children...)
}

func Rotate(dx, dy, dz interface{}, children ...openscad.Stmt) *openscad.Rotate {
	return openscad.NewRotate(dx, dy, dz, children...)
}
