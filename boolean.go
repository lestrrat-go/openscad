package openscad

import (
	"fmt"
	"io"
)

type noArgBlock struct {
	name     string
	children []Stmt
}

type Union struct{ noArgBlock }
type Difference struct{ noArgBlock }
type Intersection struct{ noArgBlock }

func (op *noArgBlock) Add(s Stmt) {
	op.children = append(op.children, s)
}

func (op *noArgBlock) Body(s ...Stmt) {
	op.children = make([]Stmt, len(s))
	copy(op.children, s)
}

func (op *noArgBlock) EmitStmt(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, `%s%s()`, ctx.Indent(), op.name)
	return emitChildren(ctx, w, op.children)
}

func NewUnion(children ...Stmt) *Union {
	return &Union{
		noArgBlock{
			name:     "union",
			children: children,
		},
	}
}

func (u *Union) Add(s Stmt) *Union {
	u.noArgBlock.Add(s)
	return u
}

func (u *Union) Body(s ...Stmt) *Union {
	u.noArgBlock.Body(s...)
	return u
}

func NewDifference(children ...Stmt) *Difference {
	return &Difference{
		noArgBlock{
			name:     "difference",
			children: children,
		},
	}
}

func (u *Difference) Add(s Stmt) *Difference {
	u.noArgBlock.Add(s)
	return u
}

func (u *Difference) Body(s ...Stmt) *Difference {
	u.noArgBlock.Body(s...)
	return u
}

func NewIntersection(children ...Stmt) *Intersection {
	return &Intersection{
		noArgBlock{
			name:     "difference",
			children: children,
		},
	}
}

func (u *Intersection) Add(s Stmt) *Intersection {
	u.noArgBlock.Add(s)
	return u
}

func (u *Intersection) Body(s ...Stmt) *Intersection {
	u.noArgBlock.Body(s...)
	return u
}
