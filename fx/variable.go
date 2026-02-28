package fx

import (
	"iter"
	"strconv"
)

type VariableKind int

const (
	VariableKindInt VariableKind = iota
	VariableKindArray
)

type Variable struct {
	Name    string
	Address int
	Kind    VariableKind
	Size    int
}

func (v *Variable) AddrAt(offset int) (address int) {
	if v.Kind == VariableKindArray {
		return v.Address + offset
	}

	return v.Address
}

type Variables struct {
	nextFreeOffset int

	all       []*Variable
	byName    map[string]*Variable
	byAddress map[int]*Variable
}

func newVariables() *Variables {
	return &Variables{
		nextFreeOffset: VariableOffset,

		all:       []*Variable{},
		byName:    make(map[string]*Variable),
		byAddress: make(map[int]*Variable),
	}
}

func (v *Variables) All() iter.Seq[*Variable] {
	return func(yield func(v *Variable) bool) {
		for _, variable := range v.all {
			if !yield(variable) {
				return
			}
		}
	}
}

func (v *Variables) ByName(name string) (variable *Variable, ok bool) {
	variable, ok = v.byName[name]
	return
}

func (v *Variables) ByAddress(address int) (variable *Variable, ok bool) {
	variable, ok = v.byAddress[address]
	return
}

func (v *Variables) allocNextOffset(size int) (offset int) {
	offset = v.nextFreeOffset
	v.nextFreeOffset += size
	return
}

func (v *Variables) Add(variable *Variable) {
	v.all = append(v.all, variable)

	if variable.Kind == VariableKindInt {
		v.byName[variable.Name] = variable
		v.byAddress[variable.Address] = variable
	} else {
		v.byName[variable.Name] = variable
		v.byAddress[variable.Address] = variable

		for i := 0; i < variable.Size; i++ {
			v.byName["__"+variable.Name+"_"+strconv.Itoa(i)] = variable
			v.byAddress[variable.Address+i] = variable
		}
	}
}

func (v *Variables) New(name string) {
	v.Add(&Variable{
		Name:    name,
		Address: v.allocNextOffset(1),
		Kind:    VariableKindInt,
		Size:    1,
	})
}

func (v *Variables) Alloc(name string, size int) {
	v.Add(&Variable{
		Name:    name,
		Address: v.allocNextOffset(size),
		Kind:    VariableKindArray,
		Size:    size,
	})
}
