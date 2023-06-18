package parser_test

import (
	"log"
	"os"
	"testing"

	"github.com/lestrrat-go/openscad"
	"github.com/lestrrat-go/openscad/parser"
)

func TestParser(t *testing.T) {
	const src = `
// Test code
cm = 10;
inch = 2.54 * cm; // one inch is 2.54 cm
global_var = 1;

function double(x) = x * x + 2 * x - 1;

module foo(a, b, c=0) {
	bar=1;
	baz="hello";
	(bar == 1) ? big() : small();
}

foo(1, 2);

render()
	foo(2, 3, 4);

translate([1, 2, 3]) {
	cube([10, 10, 10]);
	cylinder(r=5, h=10);
}
`
	stmts, err := parser.Parse([]byte(src))
	log.Printf("%s", err)
	log.Printf("%#v", stmts)
	openscad.Emit(stmts, os.Stdout)
}

func TestThreads(t *testing.T) {
	src, err := os.ReadFile("testdata/threads.scad")
	if err != nil {
		t.Fatal(err)
	}

	stmts, err := parser.Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	openscad.Emit(stmts, os.Stdout)
}
