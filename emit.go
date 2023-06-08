package openscad

import (
	"context"
	"fmt"
	"io"
	"reflect"
)

const (
	asExpr = iota
	asStmt
)

type emitContext struct{}

func emitExpr(ctx context.Context, w io.Writer, v interface{}) error {
	return emitValue(context.WithValue(ctx, emitContext{}, asExpr), w, v)
}

func emitAny(ctx context.Context, w io.Writer, v interface{}) error {
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Slice:
		fmt.Fprint(w, "[")
		for i := 0; i < rv.Len(); i++ {
			if i > 0 {
				fmt.Fprintf(w, ", ")
			}
			if err := emitValue(ctx, w, rv.Index(i).Interface()); err != nil {
				return err
			}
		}
		fmt.Fprint(w, "]")
	default:
		_, err := fmt.Fprintf(w, "%#v", v)
		if err != nil {
			return err
		}
	}
	return nil
}

func emitValue(ctx context.Context, w io.Writer, v interface{}) error {
	emitCtx := ctx.Value(emitContext{})
	switch emitCtx {
	case nil, asExpr:
		if e, ok := v.(Expr); ok {
			return e.EmitExpr(ctx, w)
		}
		return emitAny(ctx, w, v)
	case asStmt:
		if e, ok := v.(Expr); ok {
			return e.EmitExpr(ctx, w)
		}
		return emitAny(ctx, w, v)
	default:
		return emitAny(ctx, w, v)
	}
}
