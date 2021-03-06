package sempass

import (
	"shanhu.io/smlvm/lexing"
	"shanhu.io/smlvm/pl/ast"
	"shanhu.io/smlvm/pl/parse"
	"shanhu.io/smlvm/pl/tast"
	"shanhu.io/smlvm/pl/types"
	"shanhu.io/smlvm/syms"
)

// allocPrepare checks if the provided types are all allocable, and insert
// implicit type casts if needed. Only literay expression list needs alloc
// prepare.
func allocPrepare(
	b *builder, toks []*lexing.Token, lst *tast.ExprList,
) *tast.ExprList {
	ret := tast.NewExprList()
	for i, tok := range toks {
		e := lst.Exprs[i]
		t := e.Type()
		if types.IsNil(t) {
			b.CodeErrorf(tok.Pos, "pl.cannotAlloc.fromNil",
				"cannot infer type from nil for %q", tok.Lit)
			return nil
		}
		if v, ok := types.NumConst(t); ok {
			e = numCastInt(b, tok.Pos, v, e)
			if e == nil {
				return nil
			}
		}
		if !types.IsAllocable(t) {
			b.CodeErrorf(tok.Pos, "pl.cannotAlloc",
				"cannot allocate for %s", t)
			return nil
		}
		ret.Append(e)
	}
	return ret
}

func define(
	b *builder, ids []*lexing.Token, expr tast.Expr, eq *lexing.Token,
) *tast.Define {
	// check count matching
	r := expr.R()
	nleft := len(ids)
	nright := r.Len()
	if nleft != nright {
		b.CodeErrorf(eq.Pos, "pl.cannotDefine.countMismatch",
			"defined %d identifers with %d expressions",
			nleft, nright,
		)
		return nil
	}

	if exprList, ok := tast.MakeExprList(expr); ok {
		exprList = allocPrepare(b, ids, exprList)
		if exprList == nil {
			return nil
		}
		expr = exprList
	}

	var ret []*syms.Symbol
	ts := expr.R().TypeList()
	for i, tok := range ids {
		s := declareVar(b, tok, ts[i], false)
		if s == nil {
			return nil
		}
		ret = append(ret, s)
	}

	return &tast.Define{Left: ret, Right: expr}
}

func buildIdentExprList(b *builder, list *ast.ExprList) (
	idents []*lexing.Token, firstError ast.Expr,
) {
	ret := make([]*lexing.Token, 0, list.Len())
	for _, expr := range list.Exprs {
		op, ok := expr.(*ast.Operand)
		if !ok {
			return nil, expr
		}
		if op.Token.Type != parse.Ident {
			return nil, expr
		}
		ret = append(ret, op.Token)
	}

	return ret, nil
}

func buildDefineStmt(b *builder, stmt *ast.DefineStmt) tast.Stmt {
	right := b.buildExpr(stmt.Right)
	if right == nil {
		return nil
	}

	idents, err := buildIdentExprList(b, stmt.Left)
	if err != nil {
		b.Errorf(ast.ExprPos(err), "left side of := must be identifier")
		return nil
	}
	ret := define(b, idents, right, stmt.Define)
	if ret == nil {
		return nil
	}
	return ret
}
