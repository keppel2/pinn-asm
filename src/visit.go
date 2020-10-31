package main

import "fmt"
import "reflect"

const indent = "  "
var ilevel = 0

func prn(s ...interface {}) {
	for i := 0; i < ilevel; i++ {
		fmt.Print(indent)
	}
	fmt.Println(s)
}

func iminus() {
	ilevel--
}

func pnode(n Node) {
	prn(reflect.TypeOf(n), n.Gpos())
	ilevel++
}

func visitVarStmt(n VarStmt) {
	for _, vd := range n.List {
		prn("var: ", vd.Value)
	}

	visitKind(n.Kind)
}

func visitField(n Field) {
	for _, vd := range n.List {
		prn("fvar: ", vd.Value)
	}

	visitKind(n.Kind)
}

func visitTypeStmt(n TypeStmt) {
	prn("type: ", n.Wl.Value)
	visitKind(n.Kind)
}

func visitForrStmt(n ForrStmt) {
	for _, e := range n.LH {
		visitExpr(e)
	}
	prn("forr op: ", n.Op)
	visitExpr(n.RH)
	visitBlockStmt(n.B)

}
func visitReturnStmt(n ReturnStmt) {
	if n.E != nil {
		visitExpr(n.E)
	}
}

func visitFuncStmt(n FuncStmt) {
	prn("fid: ", n.Wl.Value)
	for _, vd := range n.PList {
		visitField(vd)
	}
	if n.K != nil {
		visitKind(n.K)
	}
	pnode(n.B)
	defer iminus()
	visitBlockStmt(n.B)
}

func visitKind(n Kind) {
	pnode(n)
	defer iminus()
	switch t := n.(type) {
	case MKind:
		visitMKind(t)
	case SlKind:
		visitSlKind(t)
	case ArKind:
		visitArKind(t)
	case SKind:
		visitSKind(t)
	}
}
func visitMKind(n MKind) {
	visitKind(n.K)
}

func visitSlKind(n SlKind) {
	visitKind(n.K)
}
func visitArKind(n ArKind) {
	visitExpr(n.Len)
	visitKind(n.K)
}
func visitSKind(n SKind) {
	prn("Skind", n.Wl.Value)
}

func visitBinaryExpr(n BinaryExpr) {
	visitExpr(n.LHS)
	prn("Op", n.op, ".")
	visitExpr(n.RHS)
}
func visitTrinaryExpr(n TrinaryExpr) {
	visitExpr(n.LHS)
	visitExpr(n.MS)
	visitExpr(n.RHS)
}

func visitCallExpr(n CallExpr) {
	visitExpr(n.ID)
	for _, v := range n.Params {
		visitExpr(v)
	}
}

func visitUnaryExpr(n UnaryExpr) {
	if n.E != nil {
		visitExpr(n.E)
	}
	prn("Uop", n.op)
}

func visitIndexExpr(n IndexExpr) {
	visitExpr(n.X)
	visitExpr(n.E)
	/*
		if n.Start != nil {
			visitExpr(n.Start)
		}
		prn("Inc", n.Inc)

		if n.End != nil {
			visitExpr(n.End)
		}
	*/
}

func visitArrayExpr(n ArrayExpr) {
	for _, e := range n.EL {
		visitExpr(e)
	}
}

func visitExpr(n Expr) {
	pnode(n)
	defer iminus()
	switch t := n.(type) {
	case TrinaryExpr:
		visitTrinaryExpr(t)
	case NumberExpr:
		prn("Number", t.Il.Value)
	case VarExpr:
		prn("Var", t.Wl.Value)
	case IndexExpr:
		visitIndexExpr(t)
	case BinaryExpr:
		visitBinaryExpr(t)
	case CallExpr:
		visitCallExpr(t)
	case UnaryExpr:
		visitUnaryExpr(t)
	case ArrayExpr:
		visitArrayExpr(t)
	}

}
func visitWhileStmt(n WhileStmt) {
	visitExpr(n.Cond)
	visitBlockStmt(n.B)
}

func visitLoopStmt(n LoopStmt) {
	visitBlockStmt(n.B)
}
func visitIfStmt(n IfStmt) {
	visitExpr(n.Cond)
	visitStmt(n.Then)
	if n.Else != nil {
		visitStmt(n.Else)
	}
}

func visitExprStmt(e ExprStmt) {
	visitExpr(e.Expr)
}

func visitAssignStmt(a AssignStmt) {
	for _, e := range a.LHSa {
		visitExpr(e)
	}
	prn("Op", a.Op)
	for _, e := range a.RHSa {
		visitExpr(e)
	}
}

func visitBlockStmt(t BlockStmt) {
	for _, s := range t.SList {
		visitStmt(s)
	}
}

func visitStmt(s Stmt) {
	pnode(s)
	defer iminus()
	switch t := s.(type) {
	case ForrStmt:
		visitForrStmt(t)
	case BlockStmt:
		visitBlockStmt(t)
	case VarStmt:
		visitVarStmt(t)
	case TypeStmt:
		visitTypeStmt(t)
	case FuncStmt:
		visitFuncStmt(t)
	case ExprStmt:
		visitExprStmt(t)
	case AssignStmt:
		visitAssignStmt(t)
	case IfStmt:
		visitIfStmt(t)
	case WhileStmt:
		visitWhileStmt(t)
	case LoopStmt:
		visitLoopStmt(t)
	case ReturnStmt:
		visitReturnStmt(t)
	}
}

func visitFile(f File) {
	pnode(f)
	defer iminus()
	for _, s := range f.SList {
		visitStmt(s)
	}
}