package fx

const VariableOffset = 1024 * 16

type Script struct {
	commands []*CommandNode

	labels    map[string]int
	symbols   map[string][]*AddressNode
	constants map[string]ExpressionNode
	macros    map[string]*Macro
	variables map[string]int
}

func newScript() *Script {
	return &Script{
		commands:  make([]*CommandNode, 0),
		labels:    make(map[string]int),
		symbols:   make(map[string][]*AddressNode),
		constants: make(map[string]ExpressionNode),
		macros:    make(map[string]*Macro),
		variables: make(map[string]int),
	}
}

func (s *Script) addVariable(varName string) {
	s.variables[varName] = VariableOffset + len(s.variables)
}

func (s *Script) addSymbol(label string, addr *AddressNode) {
	if _, ok := s.symbols[label]; !ok {
		s.symbols[label] = make([]*AddressNode, 0, 1)
	}

	s.symbols[label] = append(s.symbols[label], addr)
}

func (s *Script) String() (str string) {
	for _, cmd := range s.commands {
		str += cmd.String() + "\n"
	}

	return str
}

func (s *Script) PC() int {
	return len(s.commands)
}

func (s *Script) Label(name string) (pc int, ok bool) {
	pc, ok = s.labels[name]

	return
}

func (s *Script) EndOfScript() (pc int) {
	return len(s.commands) + 1
}

func (s *Script) Const(name string) (v any, ok bool) {
	v, ok = s.constants[name]
	return
}

func (s *Script) Commands() []*CommandNode {
	return s.commands
}

func (s *Script) Labels() map[string]int {
	return s.labels
}

func (s *Script) Symbols() map[string][]*AddressNode {
	return s.symbols
}

func (s *Script) Variables() map[string]int {
	return s.variables
}
