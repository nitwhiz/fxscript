package fx

type CommandType int

type Identifier int

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
	CmdPush
	CmdPop
	CmdCall
	CmdRet
	CmdGoto
	CmdSet
	CmdJumpIf

	UserCommandOffset = 100
)
