package fx

import (
	"fmt"
	"strings"
)

type ExpressionNode interface {
	exprNode()
}

type CommandNode struct {
	Type CommandType
	Args []ExpressionNode
}

type AddressNode struct {
	Address int
}

type FloatNode struct {
	Value float64
}

type IntegerNode struct {
	Value int
}

type IdentifierNode struct {
	Identifier Identifier
}

type StringNode struct {
	Value string
}

type UnaryOpNode struct {
	Operator *Token
	Expr     ExpressionNode
}

type BinaryOpNode struct {
	Left     ExpressionNode
	Operator *Token
	Right    ExpressionNode
}

func (n *FloatNode) exprNode()      {}
func (n *IntegerNode) exprNode()    {}
func (n *IdentifierNode) exprNode() {}
func (n *StringNode) exprNode()     {}
func (n *AddressNode) exprNode()    {}
func (n *BinaryOpNode) exprNode()   {}
func (n *UnaryOpNode) exprNode()    {}

func (n *CommandNode) String() string {
	if len(n.Args) == 0 {
		return fmt.Sprintf("CMD(%02d)", n.Type)
	}

	args := make([]string, len(n.Args))

	for i, arg := range n.Args {
		args[i] = fmt.Sprintf("%v", arg)
	}

	argStr := strings.Join(args, ", ")

	return fmt.Sprintf("CMD(%02d, %s)", n.Type, argStr)
}

func (n *FloatNode) String() string {
	return fmt.Sprintf("FLOAT(%f)", n.Value)
}

func (n *IntegerNode) String() string {
	return fmt.Sprintf("INT(%d)", n.Value)
}

func (n *IdentifierNode) String() string {
	return fmt.Sprintf("IDENT(%d)", n.Identifier)
}

func (n *StringNode) String() string {
	return fmt.Sprintf("STRING(\"%s\")", n.Value)
}

func (n *AddressNode) String() string {
	return fmt.Sprintf("ADDRESS(%d)", n.Address)
}

func (n *BinaryOpNode) String() string {
	return fmt.Sprintf("BINARY(%s, %s, %s)", n.Left, n.Operator, n.Right)
}

func (n *UnaryOpNode) String() string {
	return fmt.Sprintf("UNARY(%s, %s)", n.Operator, n.Expr)
}
