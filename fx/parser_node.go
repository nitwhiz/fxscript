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

type LabelNode struct {
	Name string
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

type ConstantNode struct {
	Name string
}

type StringNode struct {
	Value string
}

type UnaryOpNode struct {
	Operator rune
	Expr     ExpressionNode
}

type BinaryOpNode struct {
	Left     ExpressionNode
	Operator rune
	Right    ExpressionNode
}

func (n *FloatNode) exprNode()      {}
func (n *IntegerNode) exprNode()    {}
func (n *IdentifierNode) exprNode() {}
func (n *ConstantNode) exprNode()   {}
func (n *StringNode) exprNode()     {}
func (n *AddressNode) exprNode()    {}
func (n *BinaryOpNode) exprNode()   {}
func (n *UnaryOpNode) exprNode()    {}
func (n *LabelNode) exprNode()      {}

func (n *CommandNode) String() string {
	args := make([]string, len(n.Args))

	for i, arg := range n.Args {
		args[i] = fmt.Sprintf("%v", arg)
	}

	argStr := strings.Join(args, ", ")

	return fmt.Sprintf("$%02d %s", n.Type, argStr)
}

func (n *FloatNode) String() string {
	return fmt.Sprintf("%f", n.Value)
}

func (n *IntegerNode) String() string {
	return fmt.Sprintf("%d", n.Value)
}

func (n *IdentifierNode) String() string {
	return fmt.Sprintf("i!%d", n.Identifier)
}

func (n *ConstantNode) String() string {
	return fmt.Sprintf("[%s]", n.Name)
}

func (n *StringNode) String() string {
	return fmt.Sprintf("\"%s\"", n.Value)
}

func (n *AddressNode) String() string {
	return fmt.Sprintf("@%d", n.Address)
}

func (n *BinaryOpNode) String() string {
	return fmt.Sprintf("(%s %s %s)", n.Left, string(n.Operator), n.Right)
}

func (n *UnaryOpNode) String() string {
	return fmt.Sprintf("%s%s", string(n.Operator), n.Expr)
}

func (n *LabelNode) String() string {
	return fmt.Sprintf("@@%s", n.Name)
}
