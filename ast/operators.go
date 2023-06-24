package ast

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Group struct {
	expr interface{}
}

func NewGroup(expr interface{}) *Group {
	return &Group{
		expr: expr,
	}
}

func (g *Group) String() string {
	var sb strings.Builder
	if err := g.EmitExpr(newEmitContext(), &sb); err != nil {
		panic(err)
	}
	return sb.String()
}

func (g *Group) EmitExpr(ctx *EmitContext, w io.Writer) error {
	fmt.Fprint(w, `(`)
	if err := emitExpr(ctx, w, g.expr); err != nil {
		return err
	}
	fmt.Fprint(w, `)`)
	return nil
}

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

func (op *UnaryOp) EmitStmt(ctx *EmitContext, w io.Writer) error {
	// A unary operator for *, #, % can be used as a statement.
	var buf bytes.Buffer
	if err := emitStmt(ctx, &buf, op.expr); err != nil {
		return err
	}

	// copy all of the whitespace characters from buf
	// to w, then emit the operator.
	for {
		r, _, err := buf.ReadRune()
		if err != nil {
			return fmt.Errorf(`unary operator %q: failed to read from buffer: %w`, op.op, err)
		}
		if !unicode.IsSpace(r) {
			fmt.Fprintf(w, `%s`, op.op)
			fmt.Fprintf(w, `%c`, r)
			if _, err := buf.WriteTo(w); err != nil {
				return fmt.Errorf(`unary operator %q: failed to copy buffer to writer: %w`, op.op, err)
			}
			return nil
		}
		fmt.Fprintf(w, "%c", r)
	}
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

func (op *BinaryOp) String() string {
	var sb strings.Builder
	if err := op.EmitExpr(newEmitContext(), &sb); err != nil {
		panic(err)
	}
	return sb.String()
}

func (op *BinaryOp) BindPrecedence() int {
	switch op.op {
	case "||":
		return 1
	case "&&":
		return 2
	case "==", "!=", "<", "<=", ">", ">=":
		return 3
	case "+", "-":
		return 4
	case "*", "/", "%":
		return 5
	case "^":
		return 6
	}
	return 0
}

func (op *BinaryOp) Op() string {
	return op.op
}

func (op *BinaryOp) Right() interface{} {
	return op.right
}

func (op *BinaryOp) Left() interface{} {
	return op.left
}

func (op *BinaryOp) EmitExpr(ctx *EmitContext, w io.Writer) error {
	if err := emitExpr(ctx, w, op.left); err != nil {
		return fmt.Errorf("failed to emit left side of binary op: %v", err)
	}
	fmt.Fprintf(w, ` %s `, op.op)
	if err := emitExpr(ctx, w, op.right); err != nil {
		return fmt.Errorf("failed to emit right side of binary op: %v", err)
	}
	return nil
}

func (op *BinaryOp) EmitStmt(ctx *EmitContext, w io.Writer) error {
	return op.EmitExpr(ctx, w)
}

func (op *BinaryOp) Rearrange(op2 *BinaryOp) *BinaryOp {
	if op.BindPrecedence() > op2.BindPrecedence() {
		return NewBinaryOp(op2.op, NewBinaryOp(op.op, op.left, op2.left), op2.right)
	}
	return op
}
