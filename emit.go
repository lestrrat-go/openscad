package openscad

import (
	"fmt"
	"io"
	"reflect"
)

func Emit(stmt Stmt, w io.Writer) error {
	ctx := &EmitContext{}
	return stmt.EmitStmt(ctx, w)
}

func emitExpr(ctx *EmitContext, w io.Writer, v interface{}) error {
	return emitValue(ctx.ForceExpr(), w, v)
}

func emitAny(ctx *EmitContext, w io.Writer, v interface{}) error {
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

func emitValue(ctx *EmitContext, w io.Writer, v interface{}) error {
	if ctx.AsExpr() {
		if e, ok := v.(Expr); ok {
			return e.EmitExpr(ctx, w)
		}
	} else if ctx.AsStmt() {
		if e, ok := v.(Stmt); ok {
			return e.EmitStmt(ctx, w)
		}
	} else {
		switch v := v.(type) {
		case Expr:
			return v.EmitExpr(ctx, w)
		case Stmt:
			return v.EmitStmt(ctx, w)
		}
	}

	return emitAny(ctx, w, v)
}
