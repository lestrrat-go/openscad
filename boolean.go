package openscad

import (
	"fmt"
	"io"
)

type Union struct {
	children []Stmt
}

func NewUnion(children ...Stmt) *Union {
	return &Union{
		children: children,
	}
}

func (u *Union) Add(s Stmt) *Union {
	u.children = append(u.children, s)
	return u
}

func (u *Union) EmitStmt(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, `%sunion()`, ctx.Indent())
	return emitChildren(ctx, w, u.children)
}
