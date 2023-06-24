Examples
========

These examples are either standalone libraries that we have collected or we have
created ourselves. Some of them depend on each other.

These files are meant to be examples of how to make the depdency management of OpenSCAD
files slightly bit easier using Go's builtin dependency management.

# How to Use github.com/lestrrat-go/openscad to manage inter-dependent OpenSCAD Projects

(NOTE TO SELF: This all works, but there are too many steps for the layman. Consider
creating tools to semi-automate this)

Prepare your library files -- files that are supposed to be included in your
main OpenSCAD file. You could programmatically create the equivalent data,
but most of you probably would just want to create a `.scad` file with your
reusable code.

This next section assumes that you are familiar with how to import Go
modules, and thus do not talk about relative placement of the files, etc.
PRs to rectify this are welcome.

## Create a Go Module (that's the unit of distribution of a Go code).

### Install Go

Please refer to official Go documentation

### Create a module directory

```shell
mkdir mylibrary
cd mylibrary

git init # presumably you will put this under version control
```

Inside the above directory, prepare the go module boilerplate.
(Note: The assumption is that this will somehow be available on GitHub or similar
options. That is why the module name below contains `github.com`)

```shell
go mod init github.com/myname/mylibrary
```

Tell the go module that you are using `github.com/letrrat-go/openscad`

```shell
go get github.com/lestrrat-go/openscad
```

### Create your OpenSCAD file, and Go adapter

Inside `mylibrary`, create your `mylibrary.scad` file.

Then, create a `mylibrary.go` file, with the following contents:

```go
package mylibrary

import (
	"embed"

	"github.com/lestrrat-go/openscad"
)

//go:embed mylibrary.scad
var src embed.FS

func init() {
	if err := openscad.RegisterFile("mylibrary.scad", openscad.WithFS(src)); err != nil {
		panic(err)
	}
}
```

The `embed` bit is needed to handle this library being included from
other places, as seen later (you can't just specify the relative path of
the file because it would be evaluated at run time)

---

The above is all you need to prepare your library. Note that this Go library
does not really handle broken OpenSCAD files too well. You should debug it first
in OpenSCAD itself.

## Write Main OpenSCAD code and Include Library

Create a directory for your main file. This should also be a Go module:

```shell
mkdir myawesomemodel
cd myawesomemodel

go mod init github.com/username/myawesomemodel
go get github.com/lestrrat-go/openscad
```

Now, since you will be referencing the library you have created earlier,
you will also want to tell Go about it:

```shell
go get github.com/username/mylibrary
```

Write your OpenSCAD file as it was a regular OpenSCAD file, and include the
library that you just created:

```openscad
include <mylibrary.scad>

module my_awesome_model() { ... }
```

And do the same as you did for `mylibrary.go`, except you should reference the
required library:

```go
package myawesomemodel

import (
	"embed"

	"github.com/lestrrat-go/openscad"
    _ "github.com/username/mylibrary"
)

//go:embed myawesomemodel.scad
var src embed.FS

func init() {
	if err := openscad.RegisterFile("myawesomemodel.scad", openscad.WithFS(src)); err != nil {
		panic(err)
	}
}
```

Now create a file in the same directory named `main.go` that will emit the amalgamated file.
The amalgamated file contains all of the code in a single file, which would
make it much easier for you to keep track and distribute later.

```go
//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/lestrrat-go/openscad/ast"
	_ "github.com/username/myawesomemodel"
)

// Executing this as `go run main.go` will output the amalgamated OpenSCAD
// code to generate the design.
func main() {
	if err := _main(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func _main() error {
	return ast.EmitFile(`myawesomemodel.scad`, os.Stdout, ast.WithAmalgamation())
}
```

### Generate The Amalgamate File

Finally, execute `main.go`, and generate the file containinig all of the dependencies.

```shell
go run main.go > /path/to/openscad-files/myawesomemodel.scad
```

Make sure to enable the auto-reload feature on OpenSCAD (found under `Design > Automatic Reload and Preview`).
Now you should be able to edit `mylibrary` and `myawesomemodel`, and have it automatically create an
amalgamated file.

You can version control the libraries just as you do your Go code, and you can share them
somewhat easier than just have a file lying around and be at the mercy of your users
knowing how to correctly download and keep track of them.

# FAQ

## Why did you have to create an entire _parser_ just for this? Couldn't you just, you know, run `sed` and replace the `include` calls?

Yes, but why would I want to create an amalgamated file that contains code using
different coding styles? (Serious answer: "I felt like it")