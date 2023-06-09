package threads_test

import (
	"io"
	"testing"

	"github.com/lestrrat-go/openscad"
	"github.com/lestrrat-go/openscad/ast"
	_ "github.com/lestrrat-go/openscad/examples/threads"
	"github.com/stretchr/testify/require"
)

func TestThreads(t *testing.T) {
	stmt, ok := openscad.Lookup("threads.scad")
	if !ok {
		t.Errorf("failed to lookup threads.scad")
	}
	require.NoError(t, ast.Emit(stmt, io.Discard), `emit should succeed`)
}
