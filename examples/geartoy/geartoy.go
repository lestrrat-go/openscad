package geartoy

import (
	"github.com/lestrrat-go/openscad"
	_ "github.com/lestrrat-go/openscad/examples/gears"
	_ "github.com/lestrrat-go/openscad/examples/joints"
	_ "github.com/lestrrat-go/openscad/examples/threads"
)

func init() {
	if err := openscad.RegisterFile("geartoy.scad"); err != nil {
		panic(err)
	}
}
