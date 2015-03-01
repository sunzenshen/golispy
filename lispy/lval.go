package lispy

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sunzenshen/go-build-your-own-lisp/mpc"
)

// ltype values for lval
const (
	lvalNumType = iota
	lvalSymType
	lvalSexprType
	lvalQexprType
	lvalErrType
)

type lval struct {
	ltype int
	num   int64   // lvalNumType
	err   string  // lvalErrType
	sym   string  // lvalSymType
	cells []*lval // lvalSexprType, lvalQexprType
}

// lvalNum creates an lval number
func lvalNum(x int64) *lval {
	v := new(lval)
	v.ltype = lvalNumType
	v.num = x
	return v
}

// lvalErr creates an lval error
func lvalErr(s string) *lval {
	v := new(lval)
	v.ltype = lvalErrType
	v.err = string(s)
	return v
}

// lvalSym creates an lval symbol
func lvalSym(s string) *lval {
	v := new(lval)
	v.ltype = lvalSymType
	v.sym = string(s)
	return v
}

// lvalSexpr creates an lval S-expression
func lvalSexpr() *lval {
	v := new(lval)
	v.ltype = lvalSexprType
	v.cells = make([]*lval, 0)
	return v
}

// lvalQexpr creates an lval Q-expression
func lvalQexpr() *lval {
	v := new(lval)
	v.ltype = lvalQexprType
	v.cells = make([]*lval, 0)
	return v
}

func (v *lval) cellCount() int {
	return len(v.cells)
}

func (v *lval) ltypeString() string {
	switch v.ltype {
	case lvalNumType:
		return "lvalNumType"
	case lvalErrType:
		return "lvalErrType"
	case lvalSymType:
		return "lvalSymType"
	case lvalSexprType:
		return "lvalSexprType"
	case lvalQexprType:
		return "lvalQexprType"
	}
	return strconv.Itoa(v.ltype)
}

func (v *lval) lvalString() string {
	switch v.ltype {
	case lvalNumType:
		return strconv.FormatInt(v.num, 10)
	case lvalErrType:
		return ("Error: " + v.err)
	case lvalSymType:
		return (v.sym)
	case lvalSexprType:
		return v.lvalExprString("(", ")")
	case lvalQexprType:
		return v.lvalExprString("{", "}")
	}
	return fmt.Sprintf("Error: lvalString() unhandled ltype %d", v.ltype)
}

func (v *lval) lvalPrint() {
	fmt.Print(v.lvalString())
}

func (v *lval) lvalPrintLn() {
	v.lvalPrint()
	fmt.Print("\n")
}

func (v *lval) lvalAdd(x *lval) {
	if x == nil {
		fmt.Println("ERROR: Failed to add lval, addition is nil")
	} else {
		v.cells = append(v.cells, x)
	}
}

func (v *lval) lvalExprString(openChar string, closeChar string) string {
	s := openChar
	for i, cell := range v.cells {
		s += cell.lvalString()
		if i < v.cellCount()-1 {
			s += " "
		}
	}
	s += closeChar
	return s
}

func (v *lval) lvalExprPrint(openChar string, closeChar string) {
	fmt.Print(v.lvalExprString(openChar, closeChar))
}

func lvalReadNum(tree mpc.MpcAst) *lval {
	x, err := strconv.ParseInt(mpc.GetContents(tree), 10, 0)
	if err != nil {
		return lvalErr("Invalid Number!")
	}
	return lvalNum(x)
}

func lvalRead(tree mpc.MpcAst) *lval {
	// If Symbol or Number, return conversion to that type
	if strings.Contains(mpc.GetTag(tree), "number") {
		return lvalReadNum(tree)
	}
	if strings.Contains(mpc.GetTag(tree), "symbol") {
		return lvalSym(mpc.GetContents(tree))
	}
	// If root (>) or S-expression, then create empty list
	var x *lval
	if mpc.GetTag(tree) == ">" {
		x = lvalSexpr()
	} else if strings.Contains(mpc.GetTag(tree), "sexpr") {
		x = lvalSexpr()
	} else if strings.Contains(mpc.GetTag(tree), "qexpr") {
		x = lvalQexpr()
	}
	// Fill the cell list with any valid expressions in the children
	for i := 0; i < mpc.GetNumChildren(tree); i++ {
		iChild := mpc.GetChild(tree, i)
		iContents := mpc.GetContents(iChild)
		if iContents == "(" ||
			iContents == ")" ||
			iContents == "{" ||
			iContents == "}" ||
			mpc.GetTag(iChild) == "regex" {
			continue
		} else {
			x.lvalAdd(lvalRead(iChild))
		}
		strconv.ParseInt(mpc.GetContents(tree), 10, 0)
	}
	return x
}

func (v *lval) lvalEvalSexpr() *lval {
	// Evaluate children
	for i, cell := range v.cells {
		v.cells[i] = cell.lvalEval()
	}
	// Error checking
	for i, cell := range v.cells {
		if cell.ltype == lvalErrType {
			return v.lvalTake(i)
		}
	}
	// Empty Expression
	if v.cellCount() == 0 {
		return v
	}
	// Single Expression
	if v.cellCount() == 1 {
		return v.lvalTake(0)
	}
	// Ensure first element is a symbol
	f := v.lvalPop(0)
	if f.ltype != lvalSymType {
		return lvalErr("S-expression does not start with symbol!")
	}
	// Call builtin with operator
	return builtin(v, f.sym)
}

func (v *lval) lvalEval() *lval {
	if v.ltype == lvalSexprType {
		return v.lvalEvalSexpr()
	}
	return v
}

func (v *lval) lvalPop(i int) *lval {
	x := v.cells[i]
	copy(v.cells[i:], v.cells[i+1:])
	v.cells[len(v.cells)-1] = nil
	v.cells = v.cells[:len(v.cells)-1]
	return x
}

func (v *lval) lvalTake(i int) *lval {
	return v.lvalPop(i)
}

func lvalJoin(x *lval, y *lval) *lval {
	for y.cellCount() > 0 {
		x.lvalAdd(y.lvalPop(0))
	}
	return x
}
