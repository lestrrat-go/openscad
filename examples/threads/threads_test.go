package threads_test

import (
	"os"
	"testing"

	"github.com/lestrrat-go/openscad"
	"github.com/lestrrat-go/openscad/ast"
	_ "github.com/lestrrat-go/openscad/examples/threads"
)

func TestThreads(t *testing.T) {
	stmt, ok := openscad.Lookup("threads.scad")
	if !ok {
		t.Errorf("failed to lookup threads.scad")
	}
	ast.Emit(stmt, os.Stdout)
}
