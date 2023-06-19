package openscad_test

import (
	"log"
	"os"
	"testing"

	"github.com/lestrrat-go/openscad"
	"github.com/lestrrat-go/openscad/ast"
)

func TestParser(t *testing.T) {
	const src = `
// Test code
include <foo.scad>
cm = 10;
inch = 2.54 * cm; // one inch is 2.54 cm
global_var = 1;

function double(x) = x * x + 2 * x - 1;
function mod(x, y) = x % y == 1;

module foo(a, b, c=0) {
	bar=1;
	baz="hello";
}

foo(1, 2);

render()
	foo(2, 3, 4);

translate([1, 2, 3]) {
	cube([10, 10, 10]);
	cylinder(r=5, h=10);
}

  points = [
    for (i=[-1:NP])
      (i<0) ? midbot :
      ((i==NP) ? midtop :
      pointarrays[floor(i/P)][i%P])
  ];


`
	stmts, err := openscad.Parse([]byte(src))
	log.Printf("%s", err)
	log.Printf("%#v", stmts)
	ast.Emit(stmts, os.Stdout)
}