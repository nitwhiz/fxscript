package fx

import (
	"fmt"
	"strings"
)

type ExpressionNode interface {
	exprNode()
}

type CommandNode struct {
	*SourceInfo
	Type CommandType
	Args []ExpressionNode
}

type AddressNode struct {
	*SourceInfo
	Address int
}

type FloatNode struct {
	*SourceInfo
	Value float64
}

type IntegerNode struct {
	*SourceInfo
	Value int
}

type IdentifierNode struct {
	*SourceInfo
	Identifier Identifier
}

type StringNode struct {
	*SourceInfo
	Value string
}

type UnaryOpNode struct {
	*SourceInfo
	Operator *Token
	Expr     ExpressionNode
}

type BinaryOpNode struct {
	*SourceInfo
	Left     ExpressionNode
	Operator *Token
	Right    ExpressionNode
}

type ArrayAccessNode struct {
	*SourceInfo
	Variable Identifier
	Index    ExpressionNode
}

func (n *FloatNode) exprNode()       {}
func (n *IntegerNode) exprNode()     {}
func (n *IdentifierNode) exprNode()  {}
func (n *StringNode) exprNode()      {}
func (n *AddressNode) exprNode()     {}
func (n *BinaryOpNode) exprNode()    {}
func (n *UnaryOpNode) exprNode()     {}
func (n *ArrayAccessNode) exprNode() {}

func (n *CommandNode) String() string {
	prefix := n.SourceInfo.String()

	if len(n.Args) == 0 {
		return fmt.Sprintf("%s CMD(%02d)", prefix, n.Type)
	}

	args := make([]string, len(n.Args))

	for i, arg := range n.Args {
		args[i] = fmt.Sprintf("%v", arg)
	}

	argStr := strings.Join(args, ", ")

	return fmt.Sprintf("%s CMD(%02d, %s)", prefix, n.Type, argStr)
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

func (n *ArrayAccessNode) String() string {
	return fmt.Sprintf("AT(%d, %s)", n.Variable, n.Index)
}
