package fx

type CommandType int

type Identifier int

func (i Identifier) IsVariable() bool {
	return i >= VariableOffset
}

func (i Identifier) RealAddress() int {
	if i.IsVariable() {
		return int(i) - VariableOffset
	}

	return int(i)
}

type CommandTypeTable map[string]CommandType

type IdentifierTable map[string]Identifier

func (p *Parser) getCommandType(name string) (CommandType, bool) {
	v, ok := p.commandTypes[name]
	return v, ok
}

func (p *Parser) getIdentifier(name string) (Identifier, bool) {
	v, ok := p.identifiers[name]
	return v, ok
}

const (
	CmdNone CommandType = iota
	CmdNop
	CmdExit
	CmdPush
	CmdPop
	CmdCall
	CmdRet
	CmdGoto
	CmdSet
	CmdJumpIf

	UserCommandOffset
)
