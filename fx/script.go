package fx

type Script struct {
	commands []*CommandNode

	labels map[string]int

	constants map[string]ExpressionNode
	macros    map[string]*Script
}

func newScript(parentScript *Script) *Script {
	s := Script{
		commands: make([]*CommandNode, 0),
	}

	if parentScript != nil {
		s.labels = parentScript.labels
		s.macros = parentScript.macros
		s.constants = parentScript.constants
	} else {
		s.labels = make(map[string]int)
		s.macros = make(map[string]*Script)
		s.constants = make(map[string]ExpressionNode)
	}

	return &s
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
