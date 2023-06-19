package shedholder

import (
	"github.com/lestrrat-go/openscad"
	_ "github.com/lestrrat-go/openscad/examples/constants"
	_ "github.com/lestrrat-go/openscad/examples/joints"
	_ "github.com/lestrrat-go/openscad/examples/threads"
)

func init() {
	if err := openscad.RegisterFile("shedholder.scad"); err != nil {
		panic(err)
	}
}
