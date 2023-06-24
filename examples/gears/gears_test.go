package gears_test

import (
	"io"
	"testing"

	"github.com/lestrrat-go/openscad"
	"github.com/lestrrat-go/openscad/ast"
	_ "github.com/lestrrat-go/openscad/examples/gears"
	"github.com/stretchr/testify/require"
)

func TestGears(t *testing.T) {
	stmt, ok := openscad.Lookup("gears.scad")
	require.True(t, ok, "lookup gears.scad should succeed")
	require.NoError(t, ast.Emit(stmt, io.Discard), `emit should succeed`)
}
