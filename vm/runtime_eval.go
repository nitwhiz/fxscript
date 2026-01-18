package vm

import (
	"github.com/nitwhiz/fxscript/fx"
)

type numeric interface {
	float64 | int
}

func add[T numeric](a, b T) T {
	return a + b
}

func sub[T numeric](a, b T) T {
	return a - b
}

func mul[T numeric](a, b T) T {
	return a * b
}

func div[T numeric](a, b T) T {
	return a / b
}

func mod[T numeric](a, b T) int {
	ia := int(a)
	ib := int(b)

	return ia % ib
}

func eq[T numeric](a, b T) int {
	if a == b {
		return 1
	}

	return 0
}

func neq[T numeric](a, b T) int {
	return 1 - eq(a, b)
}

func lt[T numeric](a, b T) int {
	if a < b {
		return 1
	}

	return 0
}

func gt[T numeric](a, b T) int {
	if a > b {
		return 1
	}

	return 0
}

func lte[T numeric](a, b T) int {
	if a <= b {
		return 1
	}

	return 0
}

func gte[T numeric](a, b T) int {
	if a >= b {
		return 1
	}

	return 0
}

func shl[T numeric](a, b T) int {
	return int(a) << int(b)
}

func shr[T numeric](a, b T) int {
	return int(a) >> int(b)
}

func and[T numeric](a, b T) int {
	return int(a) & int(b)
}

func or[T numeric](a, b T) int {
	return int(a) | int(b)
}

func xor[T numeric](a, b T) int {
	return int(a) ^ int(b)
}

func evalOp[T numeric](op fx.TokenType, a, b T) (v T, err error) {
	switch op {
	case fx.ADD:
		v = add(a, b)
		break
	case fx.SUB:
		v = sub(a, b)
		break
	case fx.MUL:
		v = mul(a, b)
		break
	case fx.DIV:
		v = div(a, b)
		break
	case fx.MOD:
		v = T(mod(a, b))
		break
	case fx.LT:
		v = T(lt(a, b))
		break
	case fx.GT:
		v = T(gt(a, b))
		break
	case fx.LTE:
		v = T(lte(a, b))
		break
	case fx.GTE:
		v = T(gte(a, b))
		break
	case fx.EQ:
		v = T(eq(a, b))
		break
	case fx.NEQ:
		v = T(neq(a, b))
		break
	case fx.SHL:
		v = T(shl(a, b))
		break
	case fx.SHR:
		v = T(shr(a, b))
		break
	case fx.AND:
		v = T(and(a, b))
		break
	case fx.OR:
		v = T(or(a, b))
		break
	case fx.INV:
		v = T(xor(a, b))
		break
	default:
		err = &fx.SyntaxError{&fx.UnknownOperatorError{op}}
		break
	}

	return
}

func (f *RuntimeFrame) evalUnaryOp(n *fx.UnaryOpNode) (v any, err error) {
	if n.Operator.Type == fx.AND {
		if v, ok := n.Expr.(*fx.IdentifierNode); ok {
			return int(v.Identifier), nil
		}
	}

	v, err = f.Eval(n.Expr)

	if err != nil {
		return
	}

	switch n.Operator.Type {
	case fx.SUB:
		switch o := v.(type) {
		case int:
			v = -o
			break
		case float64:
			v = -o
			break
		}
	case fx.MUL:
		switch o := v.(type) {
		case fx.Identifier:
			v = f.Get(o)
			break
		case int:
			v = f.Get(fx.Identifier(o))
			break
		case float64:
			v = f.Get(fx.Identifier(int(o)))
			break
		}
	case fx.INV:
		switch o := v.(type) {
		case int:
			v = ^o
			break
		case float64:
			v = ^int(o)
			break
		}
	case fx.EXCL:
		switch o := v.(type) {
		case int, float64:
			if o == 0 {
				v = 1
			} else {
				v = 1
			}
			break
		}
	default:
		err = &fx.SyntaxError{&fx.UnknownOperatorError{n.Operator.Type}}
		break
	}

	return
}

func (f *RuntimeFrame) evalBinaryOp(n *fx.BinaryOpNode) (any, error) {
	left, err := f.Eval(n.Left)

	if err != nil {
		return 0, err
	}

	right, err := f.Eval(n.Right)

	if err != nil {
		return 0, err
	}

	var ok bool

	var iRight *int
	var iLeft *int

	var fRight *float64
	var fLeft *float64

	iiLeft, ok := left.(int)

	if ok {
		iLeft = &iiLeft
	}

	iiRight, ok := right.(int)

	if ok {
		iRight = &iiRight
	}

	ffLeft, ok := left.(float64)

	if ok {
		fLeft = &ffLeft
	}

	ffRight, ok := right.(float64)

	if ok {
		fRight = &ffRight
	}

	if iLeft != nil && iRight != nil {
		return evalOp(n.Operator.Type, *iLeft, *iRight)
	}

	if fLeft != nil && fRight != nil {
		return evalOp(n.Operator.Type, *fLeft, *fRight)
	}

	if iLeft != nil && fRight != nil {
		return evalOp(n.Operator.Type, float64(*iLeft), *fRight)
	}

	if fLeft != nil && iRight != nil {
		return evalOp(n.Operator.Type, *fLeft, float64(*iRight))
	}

	return 0, &fx.RuntimeError{&fx.UnexpectedBinaryOpError{left, right}}
}

func (f *RuntimeFrame) Eval(node fx.ExpressionNode) (v any, err error) {
	switch n := node.(type) {
	case *fx.BinaryOpNode:
		v, err = f.evalBinaryOp(n)
		break
	case *fx.UnaryOpNode:
		v, err = f.evalUnaryOp(n)
		break
	case *fx.StringNode:
		v = n.Value
		break
	case *fx.IntegerNode:
		v = n.Value
		break
	case *fx.FloatNode:
		v = n.Value
		break
	case *fx.IdentifierNode:
		v = f.Get(n.Identifier)
		break
	case *fx.AddressNode:
		v = n.Address
		break
	}

	return
}
