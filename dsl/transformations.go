package dsl

import "github.com/lestrrat-go/openscad"

func Translate(x, y, z interface{}, children ...openscad.Stmt) *openscad.Translate {
	return openscad.NewTranslate(x, y, z, children...)
}

func Rotate(dx, dy, dz interface{}, children ...openscad.Stmt) *openscad.Rotate {
	return openscad.NewRotate(dx, dy, dz, children...)
}
