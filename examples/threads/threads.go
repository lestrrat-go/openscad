// Based on code from https://github.com/rcolyer/threads-scad

package threads

import (
	"embed"

	"github.com/lestrrat-go/openscad"
)

//go:embed threads.scad
var src embed.FS

func init() {
	if err := openscad.RegisterFile("threads.scad", openscad.WithFS(src)); err != nil {
		panic(err)
	}
}
