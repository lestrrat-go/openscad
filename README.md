openscad
========

This is a PoC for a shim over the OpenSCAD language. 

Currently it has a half-baked OpenSCAD parser and printer, as well as building blocks to
generate OpenSCAD code programmatically.

# Parser

Currently the parser does not report errors well (especially in the lexer),
because I was being lazy during initial development.

```go
stmt, err := ast.Parse([]byte(...OpenSCAD source code...))
```

Parsed code does not contain comments.

You can output this code back by using one of the `Emit` functions:

```go
ast.Emit(stmt, os.Stdout) // emits to stdout
```

# Amalgamation

One of the goals of this library is to make (re)distribution of OpenSCAD code.
Because OpenSCAD does not by itself have a way to handle file dependencies,
it becomes increasingly harder to control these dependencies both for development
and distribution.

This library can make this slightly easier by allowing you to create amalgamated
files, where all of the related source code is provided as a single file.

For example, suppose you have an OpenSCAD code like the following:

```openscad
// main.scad
include <foo.scad>
```

Then you could register this `main.scad` file and the `foo.scad` file in this
library:

```go
openscad.RegisterFile(`main.scad`)
opnescad.RegisterFile(`foo.scad`)
```

And then you could generate the amalgamated version using the following code

```go
stmt, _ := openscad.Lookup(`main.scad`)
ast.Emit(stmt, os.Stdout, ast.WithAmalgamation())
```
