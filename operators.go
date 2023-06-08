package openscad

import (
	"context"
	"fmt"
	"io"
)

type BinaryOp struct {
	op    string
	left  interface{}
	right interface{}
}

func NewBinaryOp(op string, left, right interface{}) *BinaryOp {
	return &BinaryOp{
		op:    op,
		left:  left,
		right: right,
	}
}

func (op *BinaryOp) EmitExpr(ctx context.Context, w io.Writer) error {
	emitValue(ctx, w, op.left)
	fmt.Fprintf(w, `%s`, op.op)
	emitValue(ctx, w, op.right)
	return nil
}
