package gears

import "github.com/lestrrat-go/openscad"

func init() {
	if err := openscad.RegisterFile("gears.scad"); err != nil {
		panic(err)
	}
}
