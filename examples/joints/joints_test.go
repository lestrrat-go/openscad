package joints_test

import (
	"io"
	"testing"

	"github.com/lestrrat-go/openscad"
	"github.com/lestrrat-go/openscad/ast"
	_ "github.com/lestrrat-go/openscad/examples/joints"
	"github.com/stretchr/testify/require"
)

func TestDovetail(t *testing.T) {
	stmt, ok := openscad.Lookup("dovetail.scad")
	require.True(t, ok, "lookup dovetail.scad should succeed")
	require.NoError(t, ast.Emit(stmt, io.Discard), `emit should succeed`)
}
