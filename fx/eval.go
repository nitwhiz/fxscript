package fx

import (
	"fmt"
)

type IdentifierValueRetriever func(Identifier) any

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

func evalOp[T numeric](op *Token, a, b T) (v T, err error) {
	switch op.Type {
	case ADD:
		v = add(a, b)
	case SUB:
		v = sub(a, b)
	case MUL:
		v = mul(a, b)
	case DIV:
		v = div(a, b)
	case PERCENT:
		v = T(mod(a, b))
	case LT:
		v = T(lt(a, b))
	case GT:
		v = T(gt(a, b))
	case LTE:
		v = T(lte(a, b))
	case GTE:
		v = T(gte(a, b))
	case EQ:
		v = T(eq(a, b))
	case NEQ:
		v = T(neq(a, b))
	case SHL:
		v = T(shl(a, b))
	case SHR:
		v = T(shr(a, b))
	case AND:
		v = T(and(a, b))
	case OR:
		v = T(or(a, b))
	case INV:
		v = T(xor(a, b))
	default:
		err = &SyntaxError{op.SourceInfo, &UnknownOperatorError{TokenType: op.Type}}
	}

	return
}

func (s *Script) evalPointer(n *UnaryOpNode, getValue IdentifierValueRetriever) (v any, ok bool, err error) {
	switch n.Operator.Type {
	case AND:
		switch e := n.Expr.(type) {
		case *IdentifierNode:
			v = int(e.Identifier)
		case *IntegerNode:
			v = e.Value
		default:
			v, err = s.Eval(n.Expr, getValue)

			switch v.(type) {
			case Identifier:
				v = int(v.(Identifier))
			case int:
				v = v.(int)
			default:
				err = &RuntimeError{n.SourceInfo, &UnresolvedSymbolError{fmt.Sprintf("%+v", v)}}
			}
		}

		ok = true

		return
	case MUL:
		switch e := n.Expr.(type) {
		case *IdentifierNode:
			v = getValue(e.Identifier)
		case *IntegerNode:
			v = getValue(Identifier(e.Value))
		default:
			v, err = s.Eval(n.Expr, getValue)

			switch v.(type) {
			case Identifier:
				v = getValue(v.(Identifier))
			case int:
				v = getValue(Identifier(v.(int)))
			default:
				err = &RuntimeError{n.SourceInfo, &UnresolvedSymbolError{fmt.Sprintf("%+v", v)}}
			}
		}

		ok = true

		return
	default:
		break
	}

	return
}

func (s *Script) evalUnaryOp(n *UnaryOpNode, getValue IdentifierValueRetriever) (v any, err error) {
	var ok bool

	v, ok, err = s.evalPointer(n, getValue)

	if err != nil || ok {
		return
	}

	if v, err = s.Eval(n.Expr, getValue); err != nil {
		return
	}

	switch n.Operator.Type {
	case SUB:
		switch o := v.(type) {
		case int:
			v = -o
		case float64:
			v = -o
		}
	case INV:
		switch o := v.(type) {
		case int:
			v = ^o
		case float64:
			v = ^int(o)
		}
	case EXCL:
		switch o := v.(type) {
		case int, float64:
			if o == 0 {
				v = 1
			} else {
				v = 1
			}
		}
	default:
		err = &SyntaxError{n.SourceInfo, &UnknownOperatorError{TokenType: n.Operator.Type}}
	}

	return
}

func (s *Script) evalBinaryOp(n *BinaryOpNode, getValue IdentifierValueRetriever) (result any, err error) {
	result = 0

	var left, right any

	if left, err = s.Eval(n.Left, getValue); err != nil {
		return
	}

	if right, err = s.Eval(n.Right, getValue); err != nil {
		return
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

	err = &RuntimeError{n.Operator.SourceInfo, &UnexpectedBinaryOpError{left, right}}

	return
}

func (s *Script) EvalArrayAccessAddress(n *ArrayAccessNode, getValue IdentifierValueRetriever) (addr int, err error) {
	index, err := s.Eval(n.Index, getValue)

	if err != nil {
		return
	}

	indexInt, ok := index.(int)

	if !ok {
		err = &RuntimeError{n.SourceInfo, &UnexpectedTypeError{fmt.Sprintf("%T", index)}}
		return
	}

	baseVarName, ok := s.variableNames[int(n.Variable)]

	if !ok {
		err = &RuntimeError{n.SourceInfo, &UnresolvedSymbolError{fmt.Sprintf("%d", n.Variable)}}
		return
	}

	if indexInt > 0 {
		addr, ok = s.variables[fmt.Sprintf("__%s_%d", baseVarName, indexInt)]

		if !ok {
			err = &RuntimeError{n.SourceInfo, &UnresolvedSymbolError{fmt.Sprintf("%d+%d", n.Variable, indexInt)}}
			return
		}
	} else {
		addr = int(n.Variable)
	}

	return
}

func (s *Script) evalArrayAccess(n *ArrayAccessNode, getValue IdentifierValueRetriever) (v any, err error) {
	addr, err := s.EvalArrayAccessAddress(n, getValue)

	if err != nil {
		return
	}

	v = getValue(Identifier(addr))

	return
}

func (s *Script) Eval(node ExpressionNode, getValue IdentifierValueRetriever) (v any, err error) {
	switch n := node.(type) {
	case *BinaryOpNode:
		v, err = s.evalBinaryOp(n, getValue)
	case *UnaryOpNode:
		v, err = s.evalUnaryOp(n, getValue)
	case *StringNode:
		v = n.Value
	case *IntegerNode:
		v = n.Value
	case *FloatNode:
		v = n.Value
	case *IdentifierNode:
		v = getValue(n.Identifier)
	case *AddressNode:
		v = n.Address
	case *ArrayAccessNode:
		v, err = s.evalArrayAccess(n, getValue)
	}

	return
}
