package openscad

import (
	"context"
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

func (u *Union) EmitStmt(ctx context.Context, w io.Writer) error {
	indent := GetIndent(ctx)
	fmt.Fprintf(w, `%sunion()`, indent)
	return emitChildren(ctx, w, u.children)
}
