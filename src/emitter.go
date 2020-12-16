package main

import "fmt"
import "reflect"
import "math/rand"
import "os"

type emitter struct {
	rMap    map[string]*mloc
	cbranch branch
	ebranch branch
	moff    int
	soff    int
	lstack  [][2]branch
	fc      bool
	fexitm  map[string]branch
	fexit   branch
	lst     Node
	st      Node
	file    *File
	p       *phys
}

func (e *emitter) checks() {
	for k, v := range e.rMap {
		if !v.check() {
			e.err(k)
		}
	}
}
func (e *emitter) clearL() {
	for k, v := range e.rMap {
		if v.fc {
			delete(e.rMap, k)
		}
	}
}

func (e *emitter) emitArrayExpr(ae *ArrayExpr) *mloc {
	ml := e.newArml(len(ae.EL))
	for key, expr := range ae.EL {
		e.assignToReg(expr)
		e.p.mov(TR3, key)
		e.iStore(TR2, TR3, ml)
	}
	return ml
}

func (e *emitter) newSlc() *mloc {
	ml := new(mloc)
	e.p.mov(TR5, 0)
	ml.init(e.fc, mlSlice)
	if ml.fc {
		e.soff++
		e.soff++
		ml.i = e.soff
		e.p.push(TR5)
		e.p.push(TR5)
	} else {
		ml.i = e.moff
		e.moff++
		e.moff++
		e.p.mov(TR6, 1)
		e.iStore(TR5, TR6, ml)
		e.p.mov(TR6, 2)
		e.iStore(TR5, TR6, ml)
	}
	return ml
}

func (e *emitter) newIntml() *mloc {
	ml := new(mloc)
	ml.init(e.fc, mlInt)
	if ml.fc {
		e.soff++
		ml.i = e.soff
	} else {
		ml.i = e.moff
		e.moff++
	}
	e.p.mov(TR5, 0)
	e.storeml(ml, TR5)
	return ml
}

func (e *emitter) newArml(len int) *mloc {
	ml := new(mloc)
	ml.init(e.fc, mlArray)
	ml.len = len //atoi(e, t.Len.(*NumberExpr).Il.Value)
	if e.fc {
		e.soff += ml.len
		ml.i = e.soff
		e.p.mov(TR2, 0)
		for i := 0; i < ml.len; i++ {
			e.p.push(TR2)
		}
	} else {
		ml.i = e.moff
		e.moff += ml.len
		e.p.mov(TR2, 0)
		for i := 0; i < ml.len; i++ {
			e.p.mov(TR3, i)

			e.iStore(TR2, TR3, ml)
		}
	}
	return ml

}

func (e *emitter) newVar(s string, k Kind) {

	if _, ok := e.rMap[s]; ok {
		e.err(s)
	}
	switch t := k.(type) {
	case *SKind:
		ml := e.newIntml()
		e.rMap[s] = ml
		e.p.mov(TR2, 0)
		e.storeml(ml, TR2)
		if t.Wl.Value == "void" {
			e.rMap[s].mlt = mlVoid
			e.rMap[s].len = -1
		}

	case *ArKind:
		ml := e.newArml(atoi(e, t.Len.(*NumberExpr).Il.Value))
		e.rMap[s] = ml
	case *SlKind:
		ml := e.newSlc()
		e.rMap[s] = ml
	default:
		e.err(s)
	}
}

func (e *emitter) resetRegs() {
	for i := TR2; i <= TR10; i++ {
		e.p.mov(i, 0)
	}

}

func (e *emitter) pushAll() {

	for i := TR2; i <= TR3; i++ {
		e.p.push(i)
	}

}
func (e *emitter) popAll() {
	for i := TR2; i >= TR3; i-- {
		e.p.pop(i)
	}
}

func (e *emitter) getaddr(m *mloc) {
	if m.fc {
		e.p.add(TR2, TSS)
	} else {
		e.p.add(TR2, TBP)
	}
}
func (e *emitter) setIndex(index regi, m *mloc) {
	e.p.lsl(index, 3)
	if m.fc {
		e.p.sub(index, moffOff(m.i))
	} else {
		e.p.add(index, moffOff(m.i))
	}
}

func (e *emitter) iStore(dest regi, index regi, m *mloc) {
	if m.mlt == mlVoid {
		if L {
			e.loadml(m, TR10)
			e.p.add(index, 1)
			e.p.emit("mov", makeReg(dest), fmt.Sprintf("%v(%v,%v,8)", 0, makeReg(TR10), makeReg(index)))
		} else {
			e.loadml(m, TR10)
			e.p.add(index, 1)
			e.p.lsl(index, 3)
			e.p.str(ATeq, dest, TR10, index)
		}
		return
	}
	if m.fc {
		if L {
			e.p.emit("mov", makeReg(dest), fmt.Sprintf("%v(%v,%v,8)", -moffOff(m.i), makeReg(TSS), makeReg(index)))
		} else {
			e.setIndex(index, m)
			e.p.str(ATeq, dest, TSS, index)
		}
	} else {
		if L {
			e.p.emit("mov", makeReg(dest), fmt.Sprintf("%v(%v,%v,8)", moffOff(m.i), makeReg(TBP), makeReg(index)))
		} else {
			e.setIndex(index, m)
			e.p.str(ATeq, dest, TBP, index)
		}
	}
}

func (e *emitter) iLoad(dest regi, index regi, m *mloc) {
	if m.mlt == mlInt {
		e.err("")
	}
	if m.mlt == mlVoid { //|| m.mlt == mlSlice {
		if L {
			e.loadml(m, TR10)
			e.p.add(index, 1)
			e.p.emit("mov", fmt.Sprintf("%v(%v,%v,8)", 0, makeReg(TR10), makeReg(index)), makeReg(dest))
		} else {
			e.loadml(m, TR10)
			e.p.add(index, 1)
			e.p.lsl(index, 3)
			e.p.ldr(ATeq, dest, TR10, index)
		}
		return
	}
	if m.fc {
		if L {
			e.p.emit("mov", fmt.Sprintf("%v(%v,%v,8)", -moffOff(m.i), makeReg(TSS), makeReg(index)), makeReg(dest))
		} else {
			e.setIndex(index, m)
			e.p.ldr(ATeq, dest, TSS, index)
		}
	} else {
		if L {
			e.p.emit("mov", fmt.Sprintf("%v(%v,%v,8)", moffOff(m.i), makeReg(TBP), makeReg(index)), makeReg(dest))
		} else {
			e.setIndex(index, m)
			e.p.ldr(ATeq, dest, TBP, index)
		}
	}
}
func (e *emitter) dString() string {
	return fmt.Sprint(e.st, reflect.TypeOf(e.st), e.rMap)
}

func (e *emitter) rangeCheck(ml *mloc) {
	if ml.mlt == mlSlice {
		return
	}
	if ml.mlt == mlVoid {
		e.p.mov(TR5, -1)
		e.iLoad(TR5, TR5, ml)
		e.p.cmp(TR2, TR5)
	} else {
		e.p.cmp(TR2, ml.len)
	}

	lab := e.clab()
	e.p.br(lab, "lt")
	ln := e.st.Gpos().Line
	e.p.mov(TR1, ln)
	e.p.emitExit()

	e.p.makeLabel(lab)
}

func (e *emitter) init(f *File) {
	if L {
		RP = "%r"
	}
	rand.Seed(42)
	e.p = new(phys)
	e.p.init()
	e.rMap = make(map[string]*mloc)
	e.fexitm = make(map[string]branch)
	e.cbranch = 1
	e.fexit = e.clab()
	e.file = f
}

func (e *emitter) clab() branch {
	rt := e.cbranch
	e.cbranch++
	return rt
}

func (e *emitter) pushloop(a, b branch) {
	e.lstack = append(e.lstack, [2]branch{a, b})
}

func (e *emitter) poploop() [2]branch {
	rt := e.lstack[len(e.lstack)-1]
	e.lstack = e.lstack[0 : len(e.lstack)-1]
	return rt
}

func (e *emitter) peekloop() [2]branch {
	return e.lstack[len(e.lstack)-1]
}

func (e *emitter) err(msg string) {
	ms := fmt.Sprintln(e.p.sb, "\n,msg,", msg, "\n", e.dString())
	fmt.Fprintln(os.Stderr, ms)
	panic("")
}

func (e *emitter) loadml(ml *mloc, r regi) {
	if ml.mlt == mlArray {
		e.err(fmt.Sprint(ml.mlt))
	}
	if ml.fc {
		e.p.ldr(ATeq, r, TSS, -moffOff(ml.i))
	} else {
		e.p.ldr(ATeq, r, TBP, moffOff(ml.i))
	}
}

func (e *emitter) storeml(ml *mloc, r regi) {
	if ml.mlt == mlArray {
		e.err("")
	}
	if ml.fc {
		e.p.str(ATeq, r, TSS, -moffOff(ml.i))
	} else {
		e.p.str(ATeq, r, TBP, moffOff(ml.i))
	}
}

func (e *emitter) loadId(v string, r regi) {
	ml, ok := e.rMap[v]
	if !ok {
		e.err(v)
	}
	e.loadml(ml, r)
}

func (e *emitter) storeInt(v string, r regi) {
	ml, ok := e.rMap[v]
	if !ok {
		e.err(v)
	}
	e.storeml(ml, r)
}

func (e *emitter) storeId(v string, r regi) {
	ml, ok := e.rMap[v]
	if !ok {
		ml = e.newIntml()
		e.rMap[v] = ml
	}

	e.storeInt(v, r)

}
func (e *emitter) doOp(dest, b regi, op string) {
	switch op {
	case "+":
		e.p.add(dest, b)
		return
	case "-":
		e.p.sub(dest, b)
		return
	case "*":
		e.p.mul(dest, b)
		return
	case "/":
		e.p.div(dest, b)
		return
	case "%":
		e.p.rem(dest, b)
		return
	case ":", "@":
		return
	default:
		e.err(op)
	}
}
func (e *emitter) condExpr(dest branch, be *BinaryExpr) {
	if be.op == "||" {
		e.condExpr(dest, be.LHS.(*BinaryExpr))
		e.condExpr(dest, be.RHS.(*BinaryExpr))
	} else if be.op == "&&" {
		lab := e.clab()
		lab2 := e.clab()
		e.condExpr(lab, be.LHS.(*BinaryExpr))
		e.p.br(lab2)
		e.p.makeLabel(lab)
		e.condExpr(dest, be.RHS.(*BinaryExpr))
		e.p.makeLabel(lab2)
	} else if be.op == "==" || be.op == "!=" || be.op == "<" || be.op == "<=" || be.op == ">" || be.op == ">=" {
		e.assignToReg(be.LHS)
		e.p.mov(TR4, TR2)
		e.assignToReg(be.RHS)
		e.p.cmp(TR4, TR2)
		bi := ""
		switch be.op {
		case "==":
			bi = "eq"
		case "!=":
			bi = "ne"
		case "<":
			bi = "lt"
		case "<=":
			bi = "le"
		case ">":
			bi = "gt"
		case ">=":
			bi = "ge"
		default:
			e.err(be.op)
		}
		e.p.br(branch(dest), bi)
		return
	} else {
		e.err(be.op)
	}

}

func (e *emitter) binaryExpr(be *BinaryExpr) *mloc {
	var rt *mloc
	lmlL := e.newIntml()

	e.assignToReg(be.LHS)
	e.storeml(lmlL, TR2)
	e.assignToReg(be.RHS)
	e.p.mov(TR3, TR2)
	e.loadml(lmlL, TR2)
	e.doOp(TR2, TR3, be.op)
	if be.op == ":" || be.op == "@" {
		rt = newSent(mlRange)
	} else {
		rt = newSent(mlInt)
	}
	return rt
}

func (e *emitter) emitFunc(f *FuncDecl) {
	e.p.flabel(f.Wl.Value)
	e.soff = 0
	e.p.mov(TSS, TSP)
	for _, field := range f.PList {
		if ark, ok := field.Kind.(*ArKind); ok {
			for _, vd2 := range field.List {

				if _, ok := e.rMap[vd2.Value]; ok {
					e.err(vd2.Value)
				}
				ml := new(mloc)
				ml.init(e.fc, mlArray)
				plen := atoi(e, ark.Len.(*NumberExpr).Il.Value)
				e.soff += plen
				ml.len = plen
				ml.i = -(f.PSize - e.soff)
				e.rMap[vd2.Value] = ml
			}
			continue

		}
		for _, vd2 := range field.List {

			if _, ok := e.rMap[vd2.Value]; ok {
				e.err(vd2.Value)
			}
			ml := new(mloc)
			ml.init(e.fc, mlInt)
			e.soff++
			ml.i = -(f.PSize - e.soff)
			e.rMap[vd2.Value] = ml
		}
	}
	e.soff = 0
	lab := e.clab()
	e.ebranch = lab
	e.emitStmt(f.B)
	e.p.makeLabel(lab)

	e.p.mov(TSP, TSS)
	e.p.emit("ret")
	e.checks()
	e.clearL()
}

func (e *emitter) assignToReg(ex Expr) *mloc {
	var rt *mloc
	e.lst = e.st
	e.st = ex
	defer func() { e.st = e.lst }()
	switch t2 := ex.(type) {
	case *ArrayExpr:
		rt = e.emitArrayExpr(t2)
		return rt
	case *NumberExpr:
		e.p.mov(TR2, atoi(e, t2.Il.Value))
		return newSent(mlInt)
	case *StringExpr:
		return newSent(mlString)
	case *VarExpr:
		rt = e.rMap[t2.Wl.Value]
		if rt.mlt != mlArray && rt.mlt != mlSlice {
			e.loadId(t2.Wl.Value, TR2)
		}
	case *BinaryExpr:
		rt = e.binaryExpr(t2)
	case *UnaryExpr:
		rt = newSent(mlInt)
		if t2.op == "-" {
			e.assignToReg(t2.E)
			e.p.mov(TR5, -1)
			e.p.mul(TR2, TR5)
		} else if t2.op == "&" {
			switch t3 := t2.E.(type) {
			case *VarExpr:
				v := t3.Wl.Value
				ml := e.rMap[v]
				e.p.mov(TR2, 0)
				e.setIndex(TR2, ml)
				e.getaddr(ml)
			case *IndexExpr:
				v := t3.X.(*VarExpr).Wl.Value
				ml := e.rMap[v]
				e.assignToReg(t3.E)
				e.rangeCheck(ml)
				e.setIndex(TR2, ml)
				e.getaddr(ml)
			}
		} else if t2.op == "*" {
			e.assignToReg(t2.E)
			e.p.ldr(ATeq, TR2, TR2)
		}
	case *TrinaryExpr:
		lab := e.clab()
		lab2 := e.clab()
		lab3 := e.clab()
		e.condExpr(lab, t2.LHS.(*BinaryExpr))
		e.p.br(lab3)
		e.p.makeLabel(lab)
		rt2 := e.assignToReg(t2.MS)
		e.p.br(lab2)
		e.p.makeLabel(lab3)
		rt3 := e.assignToReg(t2.RHS)
		if !rt2.typeOk(rt3) {
			e.err("")
		}
		e.p.makeLabel(lab2)
		rt = rt2

	case *CallExpr:
		e.emitCall(t2)
		e.p.mov(TR2, TR1)
		rt = newSent(mlInt)
	case *IndexExpr:
		v := t2.X.(*VarExpr).Wl.Value
		ml := e.rMap[v]
		ert := e.assignToReg(t2.E)
		if ert == nil {
			e.err(v)
		}
		if ert.mlt == mlRange {
			rt = ml
			break
		}
		if ml.mlt == mlSlice {
			e.p.mov(TR5, 1)
			e.p.mov(TR3, TR2)
			e.p.lsl(TR3, 3)
			e.iLoad(TR2, TR2, ml)
			e.p.ldr(ATeq, TR2, TR2, TR3)
			break

		}
		e.rangeCheck(ml)
		e.iLoad(TR2, TR2, ml)
		rt = newSent(mlInt)
	default:
		e.err("")
	}
	return rt
}

func (e *emitter) emitCall(ce *CallExpr) {
	e.st = ce
	ID := ce.ID.(*VarExpr).Wl.Value
	if ff, ok := fmap[ID]; ok {
		ff(e, ce)
		return
	}

	if ID == "print" || ID == "println" || ID == "printchar" {
		didPrint = true
	}
	fn := fmake(ID)
	fun := e.file.getFunc(ID)
	if fun == nil {
		e.err(ID)
	}
	if len(ce.Params) != fun.PCount {
		e.err(ID)
	}

	e.pushAll()
	e.p.push(TSS)
	if !L {
		e.p.push(LR)
	}
	ssize := fun.PSize

	for k, v := range ce.Params {
		if v, ok := v.(*StringExpr); ok {
			sl := len(v.W.Value)
			for _, r := range revString(v.W.Value) {
				e.p.mov(TR2, int(r))
				e.p.push(TR2)
			}
			e.p.mov(TR10, sl)
			ssize = sl
			break
		}
		var kind Kind
		if len(fun.NTlist) != 0 {
			kind = fun.NTlist[k].K
		}
		if ie, ok := v.(*VarExpr); ok && e.rMap[ie.Wl.Value].len > 0 {
			if atoi(e, kind.(*ArKind).Len.(*NumberExpr).Il.Value) != e.rMap[ie.Wl.Value].len {
				e.err(ID)
			}
			ml := e.rMap[ie.Wl.Value]
			for i := ml.len - 1; i >= 0; i-- {
				e.p.mov(TR2, i)
				e.iLoad(TR2, TR2, ml)
				e.p.push(TR2)
			}
		} else {
			if kind != nil {
				if _, ok := kind.(*SKind); !ok {
					e.err(ID)
				}
			}
			e.assignToReg(v)
			e.p.push(TR2)
		}
	}

	if L {
		e.p.emit("call", fn)
	} else {
		e.p.emit("bl", fn)
	}
	e.p.add(TSP, moffOff(ssize))
	if !L {
		e.p.pop(LR)
	}
	e.p.pop(TSS)

	e.popAll()

}

func (e *emitter) emitStmt(s Stmt) {
	e.st = s
	e.p.emit("//")
	switch t := s.(type) {
	case *ExprStmt:
		e.assignToReg(t.Expr)
	case *BlockStmt:
		for _, s := range t.SList {
			e.emitStmt(s)
		}
	case *ContinueStmt:
		e.p.br(e.peekloop()[0])
	case *BreakStmt:
		e.p.br(e.peekloop()[1])
	case *LoopStmt:
		lab := e.clab()
		lab2 := e.clab()
		e.p.makeLabel(lab)
		e.pushloop(lab, lab2)
		e.emitStmt(t.B)
		e.p.br(lab)
		e.p.makeLabel(lab2)
		e.poploop()
	case *WhileStmt:
		lab := e.clab()
		lab2 := e.clab()
		lab3 := e.clab()
		e.p.makeLabel(lab)
		e.pushloop(lab, lab2)
		e.condExpr(lab3, t.Cond.(*BinaryExpr))
		e.p.br(lab2)
		e.p.makeLabel(lab3)
		e.emitStmt(t.B)
		e.p.br(lab)
		e.p.makeLabel(lab2)
		e.poploop()

	case *IfStmt:
		lab := e.clab()
		if t.Else == nil {
			lab2 := e.clab()
			e.condExpr(lab2, t.Cond.(*BinaryExpr))
			e.p.br(lab)
			e.p.makeLabel(lab2)
			e.emitStmt(t.Then)
		} else {
			lab2 := e.clab()
			lab3 := e.clab()
			e.condExpr(lab2, t.Cond.(*BinaryExpr))
			e.p.br(lab3)
			e.p.makeLabel(lab2)
			e.emitStmt(t.Then)
			e.p.br(lab)
			e.p.makeLabel(lab3)
			e.emitStmt(t.Else)
		}
		e.p.makeLabel(lab)

	case *ReturnStmt:
		if t.E != nil {
			e.assignToReg(t.E)
			e.p.mov(TR1, TR2)
		} else {
			e.p.mov(TR1, 5)
		}
		if L {
		} else {
		}
		e.p.br(e.ebranch)
	case *AssignStmt:
		lh := t.LHSa[0]
		switch lh2 := lh.(type) {
		case *UnaryExpr:
			if lh2.op != "*" {
				e.err(lh2.op)
			}
			e.assignToReg(lh2.E)
			e.p.mov(TR3, TR2)
			e.assignToReg(t.RHSa[0])
			e.p.str(ATeq, TR2, TR3)
		case *VarExpr:
			id := lh2.Wl.Value
			if t.Op == ":=" && e.rMap[id] != nil {
				e.err(id)
			}
			if t.Op == "=" && e.rMap[id] == nil {
				e.err(id)
			}
			if t.Op == "+=" || t.Op == "-=" || t.Op == "/=" || t.Op == "*=" || t.Op == "%=" || t.Op == "++" || t.Op == "--" {
				e.loadId(id, TR3)
				if t.Op[1:2] == "=" {
					e.assignToReg(t.RHSa[0])
				} else {
					e.p.mov(TR2, 1)
				}
				e.doOp(TR3, TR2, t.Op[0:1])
				e.storeInt(id, TR3)
				return
			}
			if ae, ok := t.RHSa[0].(*BinaryExpr); t.Op == ":=" && ok && (ae.op == ":" || ae.op == "@") {
				e.p.mov(TR10, THP)
				e.storeId(id, TR10)
				e.rMap[id].mlt = mlVoid

				e.assignToReg(ae.LHS)
				e.p.mov(TR3, TR2)

				e.assignToReg(ae.RHS)
				e.p.mov(TR9, TR2)
				e.p.sub(TR9, TR3)
				if ae.op == "@" {
					e.p.add(TR9, 1)
				}

				e.p.mov(TR8, -1)
				e.iStore(TR9, TR8, e.rMap[id]) //len

				e.p.add(TR9, 1) // Add len at start
				e.p.lsl(TR9, 3)
				e.p.add(THP, TR9)
				e.p.mov(TR9, 0)
				lab := e.clab()
				e.p.makeLabel(lab)
				e.p.mov(TR8, TR9)
				e.iStore(TR3, TR9, e.rMap[id])
				e.p.mov(TR9, TR8)
				e.p.add(TR9, 1)
				e.p.add(TR3, 1)
				e.p.cmp(TR3, TR2)
				e.p.br(lab, "le")
				return
			}
			if ae, ok := t.RHSa[0].(*CallExpr); ok && ae.ID.(*VarExpr).Wl.Value == "malloc" {
				e.assignToReg(t.RHSa[0])
				e.storeId(id, TR2)
				e.rMap[id].mlt = mlVoid
				e.p.mov(TR3, -1)
				e.iStore(TR2, TR3, e.rMap[id])
				return
			}

			ml := e.assignToReg(t.RHSa[0])
			if e.rMap[id] != nil && e.rMap[id].mlt == mlSlice {
				mls := e.rMap[id]
				e.p.mov(TR4, TR3)
				e.p.sub(TR4, TR2)
				e.p.mov(TR5, 0)
				e.iStore(TR4, TR5, mls)
				e.p.mov(TR5, 1)
				e.setIndex(TR2, ml)
				e.getaddr(ml)
				e.iStore(TR2, TR5, mls)
				return
			}
			if ml != nil && ml.mlt == mlArray {
				if e.rMap[id] != nil && !e.rMap[id].typeOk(ml) {
					e.err(id)
				}
				e.rMap[id] = ml
				return
			}

			e.storeId(id, TR2)

		case *IndexExpr:
			if t.Op == "+=" || t.Op == "-=" || t.Op == "/=" || t.Op == "*=" || t.Op == "%=" || t.Op == "++" || t.Op == "--" {
				e.assignToReg(lh2)
				e.p.mov(TR3, TR2)
				if t.Op[1:2] == "=" {
					e.assignToReg(t.RHSa[0])
				} else {
					e.p.mov(TR2, 1)
				}
				e.doOp(TR3, TR2, t.Op[0:1])
			} else {
				e.assignToReg(t.RHSa[0])
				e.p.mov(TR3, TR2)
			}

			v := lh2.X.(*VarExpr).Wl.Value
			ml := e.rMap[v]

			e.assignToReg(lh2.E)
			e.rangeCheck(ml)
			e.iStore(TR3, TR2, ml)
		default:
			e.err("")
		}

	case *VarStmt:
		for _, v := range t.List {
			e.newVar(v.Value, t.Kind)
		}
	case *ForStmt:
		if t.Inits != nil {
			if rs, ok := t.Inits.(*AssignStmt); ok {
				if ue, ok := rs.RHSa[0].(*UnaryExpr); ok && ue.op == "range" {
					var iter, key *mloc
					if len(rs.LHSa) == 2 {
						key = e.rMap[rs.LHSa[0].(*VarExpr).Wl.Value]
						iter = e.rMap[rs.LHSa[1].(*VarExpr).Wl.Value]
					} else {
						iter = e.rMap[rs.LHSa[0].(*VarExpr).Wl.Value]
					}
					var ml *mloc
					ml = e.assignToReg(ue.E)
					lab := e.clab()
					lab2 := e.clab()
					e.pushloop(lab, lab2)
					e.p.mov(TR10, 0)
					e.p.makeLabel(lab)
					if key != nil {
						e.storeml(key, TR10)
					}
					e.iLoad(TR9, TR10, ml)
					e.storeml(iter, TR9)
					e.emitStmt(t.B)
					e.p.add(TR10, 1)
					e.p.cmp(TR10, ml.len)
					e.p.br(lab, "lt")
					e.p.makeLabel(lab2)
					return
				}
			}

			e.emitStmt(t.Inits)
		}

		lab := e.clab()
		lab2 := e.clab()
		lab3 := e.clab()
		lab4 := e.clab()
		lab5 := e.clab()
		e.p.makeLabel(lab)
		e.pushloop(lab, lab2)
		e.p.br(lab3)
		e.p.makeLabel(lab5)
		if t.Loop != nil {
			e.emitStmt(t.Loop)
		}
		e.p.makeLabel(lab3)

		if t.E != nil {
			e.condExpr(lab4, t.E.(*BinaryExpr))
		} else {
			e.p.br(lab4)
		}
		e.p.br(lab2)
		e.p.makeLabel(lab4)
		e.emitStmt(t.B)
		e.p.br(lab5)

		e.p.makeLabel(lab2)
		e.poploop()

	default:
		e.err("")

	}

}

func (e *emitter) emitDefines() {
	if L {
		for r := TR1; r <= TSS; r++ {
			e.p.padd("#define " + rs[r] + " " + fmt.Sprintf("%v%v", RP, irs[r]) + "\n")
		}
	} else {
		for r := TR1; r <= TSS; r++ {
			e.p.padd("#define " + rs[r] + " " + fmt.Sprintf("%v%v", RP, ars[r]) + "\n")
		}
	}
}

func (e *emitter) emitF() {
	e.p.emitDefines()
	if L {
		e.p.padd(".global _main\n")
		e.p.label("_main")
		e.p.emitR("pop", TMAIN)
		e.p.emitR("push", TMAIN)
	} else {
		e.p.padd(".global main\n")
		e.p.label("main")
		e.p.mov(TMAIN, LR)
	}
	e.p.mov(TSP, SP)
	e.p.sub(TSP, 0x100)
	e.p.mov(TSS, TSP)
	e.p.mov(TBP, TSP)
	e.p.sub(TBP, 0x1000)
	e.p.mov(THP, TBP)
	e.p.sub(THP, 0x1000)
	lab := e.clab()
	e.ebranch = lab
	for _, s := range e.file.SList {
		e.emitStmt(s)
	}
	e.p.mov(TR1, 0)
	e.p.makeLabel(lab)
	e.p.emit("ret")
	e.checks()
	e.fc = true
	for _, s := range e.file.FList {
		if s.B != nil {
			e.emitFunc(s)
		}
	}
	if didPrint {
		e.p.emitPrint(e)
	}
	e.p.makeLabel(e.fexit)
	//e.p.mov(TR1, 7)
	e.p.emit("ret")
}
