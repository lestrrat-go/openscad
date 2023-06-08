openscad
========

This is a PoC for a shim over the OpenSCAD language. It is mainly a simple code generator,
but the idea is to overcome the fact that OpenSCAD does not have a dependency management
system to easily reuse existing code by modularizing OpenSCAD source as Go code, and then
use Go's tools to package and resolve the dependencies.

```go
func Example() {
	width := dsl.Variable("width").Value(30)
	stmts := dsl.Stmts(
		width,
		dsl.Module("foobar").
			Parameters(width).
			Actions(
				dsl.Rotate(
					0, 180, 0,
					dsl.Translate(
						dsl.List(10, 10, 10),
						dsl.Cube(width, 40, 5).Fn(24),
						dsl.Cube(5, 40, width),
						dsl.Cylinder(10, 5, 15).Fa(12),
					),
				),
			),
		dsl.Call("foobar"),
	)

	stmts.Emit(context.Background(), os.Stdout)
	//OUTPUT:
	// width=30;
	//
	// module foobar(width=30)
	// {
	//   rotate([0, 180, 0])
	//     translate([10, 10, 10])
	//     {
	//       cube([width, 40, 5], $fn=24);
	//       cube([5, 40, width]);
	//       cylinder(h=10, r1=5, r2=15, $fa=12);
	//     }
	// }
	// foobar();
}
```
