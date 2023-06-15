package openscad

import (
	"fmt"
	"io"
)

type UnaryOp struct {
	op   string
	expr interface{}
}

func NewUnaryOp(op string, expr interface{}) *UnaryOp {
	return &UnaryOp{
		op:   op,
		expr: expr,
	}
}

func (op *UnaryOp) EmitExpr(ctx *EmitContext, w io.Writer) error {
	fmt.Fprintf(w, `%s`, op.op)
	if err := emitExpr(ctx, w, op.expr); err != nil {
		return err
	}
	return nil
}

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

func (op *BinaryOp) EmitExpr(ctx *EmitContext, w io.Writer) error {
	if ctx.IsNestedBinaryOp() && op.op != "*" {
		fmt.Fprint(w, `(`)
		defer fmt.Fprint(w, `)`)
	}
	ctx = ctx.WithNestedBinaryOp(true)
	if err := emitExpr(ctx, w, op.left); err != nil {
		return fmt.Errorf("failed to emit left side of binary op: %v", err)
	}
	fmt.Fprintf(w, `%s`, op.op)
	if err := emitExpr(ctx, w, op.right); err != nil {
		return fmt.Errorf("failed to emit right side of binary op: %v", err)
	}
	return nil
}

func (op *BinaryOp) EmitStmt(ctx *EmitContext, w io.Writer) error {
	return op.EmitExpr(ctx, w)
}
