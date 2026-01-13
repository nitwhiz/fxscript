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

func mod[T numeric](a, b T) T {
	ia := int(a)
	ib := int(b)

	return T(ia % ib)
}

func evalOp[T numeric](op rune, a, b T) T {
	switch op {
	case fx.OpAdd:
		return add(a, b)
	case fx.OpSub:
		return sub(a, b)
	case fx.OpMul:
		return mul(a, b)
	case fx.OpDiv:
		return div(a, b)
	case fx.OpMod:
		return mod(a, b)
	default:
		return 0
	}
}

func (f *RuntimeFrame) dispatchUnaryOpEval(n *fx.UnaryOpNode) any {
	operand := f.eval(n.Expr)

	switch n.Operator {
	case fx.OpSub:
		switch operand.(type) {
		case int:
			return -operand.(int)
		case float64:
			return -operand.(float64)
		default:
			break
		}
	default:
		break
	}

	return operand
}

func (f *RuntimeFrame) dispatchBinaryOpEval(n *fx.BinaryOpNode) any {
	left := f.eval(n.Left)
	right := f.eval(n.Right)

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
		return evalOp(n.Operator, *iLeft, *iRight)
	}

	if fLeft != nil && fRight != nil {
		return evalOp(n.Operator, *fLeft, *fRight)
	}

	if iLeft != nil && fRight != nil {
		return evalOp(n.Operator, float64(*iLeft), *fRight)
	}

	if fLeft != nil && iRight != nil {
		return evalOp(n.Operator, *fLeft, float64(*iRight))
	}

	return 0
}

func (f *RuntimeFrame) eval(node fx.ExpressionNode) any {
	switch n := node.(type) {
	case *fx.BinaryOpNode:
		return f.dispatchBinaryOpEval(n)
	case *fx.UnaryOpNode:
		return f.dispatchUnaryOpEval(n)
	case *fx.StringNode:
		return n.Value
	case *fx.IntegerNode:
		return n.Value
	case *fx.FloatNode:
		return n.Value
	case *fx.IdentifierNode:
		return int(n.Identifier)
	case *fx.VariableNode:
		return f.Get(n.Variable)
	case *fx.AddressNode:
		return n.Address
	case *fx.FlagNode:
		return int(n.Flag)
	default:
		return 0
	}
}
