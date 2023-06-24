package gears

import (
	"embed"

	"github.com/lestrrat-go/openscad"
)

//go:embed gears.scad
var src embed.FS

func init() {
	if err := openscad.RegisterFile("gears.scad", openscad.WithFS(src)); err != nil {
		panic(err)
	}
}
