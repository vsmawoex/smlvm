package g8

import (
	"fmt"

	"e8vm.io/e8vm/g8/ast"
	"e8vm.io/e8vm/g8/ir"
	"e8vm.io/e8vm/g8/parse"
	"e8vm.io/e8vm/g8/tast"
	"e8vm.io/e8vm/g8/types"
	"e8vm.io/e8vm/lex8"
)

func buildConst(b *builder, c *tast.Const) *ref {
	if _, ok := types.NumConst(c.T); ok {
		// untyped consts are just types.
		return newRef(c.T, nil)
	}

	if t, ok := c.T.(types.Basic); ok {
		v := c.ConstValue.(int64)
		switch t {
		case types.Int, types.Uint:
			return newRef(c.T, ir.Num(uint32(v)))
		case types.Int8, types.Uint8, types.Bool:
			return newRef(c.T, ir.Byt(uint8(v)))
		default:
			panic("other basic types not supported yet")
		}
	}

	if c.T == types.String {
		s := c.ConstValue.(string)
		ret := b.newTemp(c.T)
		b.b.Arith(ret.IR(), nil, "makeStr", b.p.NewString(s))
		return ret
	}

	panic("other const types not supported")
}

func buildField(b *builder, this ir.Ref, field *types.Field) *ref {
	retIR := ir.NewAddrRef(
		this,
		field.T.Size(),
		field.Offset(),
		types.IsByte(field.T),
		true,
	)
	return newAddressableRef(field.T, retIR)
}

func buildIdent(b *builder, id *tast.Ident) *ref {
	s := id.Symbol
	switch s.Type {
	case tast.SymVar:
		return s.Obj.(*objVar).ref
	case tast.SymFunc:
		v := s.Obj.(*objFunc)
		if !v.isMethod {
			return v.ref
		}
		if b.this == nil {
			panic("this missing")
		}
		return newRecvRef(v.Type().(*types.Func), b.this, v.IR())
	case tast.SymConst:
		return s.Obj.(*objConst).ref
	case tast.SymType, tast.SymStruct:
		return s.Obj.(*objType).ref
	case tast.SymField:
		v := s.Obj.(*objField)
		return buildField(b, b.this.IR(), v.Field)
	case tast.SymImport:
		return s.Obj.(*objImport).ref
	default:
		b.Errorf(id.Token.Pos, "todo: token type: %s", tast.SymStr(s.Type))
		return nil
	}
}

func buildConstIdent(b *builder, id *tast.Ident) *ref {
	s := id.Symbol
	switch s.Type {
	case tast.SymConst:
		return s.Obj.(*objConst).ref
	case tast.SymType, tast.SymStruct:
		return s.Obj.(*objType).ref
	case tast.SymImport:
		return s.Obj.(*objImport).ref
	default:
		b.Errorf(id.Token.Pos, "%s is a %s, not a const",
			id.Token.Lit, tast.SymStr(s.Type),
		)
		return nil
	}
}

func genExpr(b *builder, expr tast.Expr) *ref {
	switch expr := expr.(type) {
	case *tast.Const:
		return buildConst(b, expr)
	case *tast.Ident:
		return buildIdent(b, expr)
	}
	panic(fmt.Errorf("genExpr not implemented for %T", expr))
}

func genConstExpr(b *builder, expr tast.Expr) *ref {
	switch expr := expr.(type) {
	case *tast.Const:
		return buildConst(b, expr)
	case *tast.Ident:
		return buildConstIdent(b, expr)
	}
	panic("bug")
}

func buildOperand2(b *builder, op *lex8.Token) *ref {
	expr := b.spass.BuildExpr(&ast.Operand{op})
	if expr == nil {
		return nil
	}
	return genExpr(b, expr)
}

func buildOperand(b *builder, op *ast.Operand) *ref {
	if op.Token.Type == parse.Keyword && op.Token.Lit == "this" {
		if b.this == nil {
			b.Errorf(op.Token.Pos, "using this out of a method function")
			return nil
		}
		return b.this
	}

	switch op.Token.Type {
	case parse.Int, parse.Char, parse.String, parse.Ident:
		return buildOperand2(b, op.Token)
	default:
		b.Errorf(op.Token.Pos, "invalid or not implemented: %d",
			op.Token.Type,
		)
		return nil
	}
}

func buildConstOperand(b *builder, op *ast.Operand) *ref {
	expr := b.spass.BuildConstExpr(&ast.Operand{op.Token})
	if expr == nil {
		return nil
	}
	return genConstExpr(b, expr)
}
