package main

type SymbolTable struct {
	static []SymbolEntity
	field []SymbolEntity
	local []SymbolEntity
	arg []SymbolEntity
	subroutine []SymbolEntity
}

type SymbolEntity struct {
	name string
	typeName string
	kind string
	number int
}

var kinds = []string{"local", "arg", "field", "static"}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		static: []SymbolEntity{},
		field: []SymbolEntity{},
		local: []SymbolEntity{},
		arg: []SymbolEntity{},
		subroutine: []SymbolEntity{},
	}
}

func (s *SymbolTable) StartSubroutine() {
	s.local = []SymbolEntity{}
	s.arg = []SymbolEntity{}
}

func (s *SymbolTable) Define(name string, typeName string, kind string) {
	number := s.VarCount(kind)
	list := s.getList(kind)
	list = append(list, SymbolEntity{name: name, typeName: typeName, kind: kind, number: number})
	switch kind {
	case "static":
		s.static = list
	case "field":
		s.field = list
	case "local":
		s.local = list
	case "arg":
		s.arg = list
	case "function", "method", "constructor":
		s.subroutine = list
	}
}

func (s *SymbolTable) VarCount(kind string) int {
	return len(s.getList(kind))
}

func (s *SymbolTable) getList(kind string) []SymbolEntity {
	switch kind {
	case "static":
		return s.static
	case "field":
		return s.field
	case "local":
		return s.local
	case "arg":
		return s.arg
	case "function", "method", "constructor":
		return s.subroutine
	}
	return nil
}

func (s *SymbolTable) KindOf(name string) string {
	for _, kind := range kinds {
		if _, ok := s.Contains(name, kind); ok {
			return kind
		}
	}
	return "none"
}

func (s *SymbolTable) Contains(name, kind string) (SymbolEntity, bool) {
	list := s.getList(kind)
	for _, e := range list {
		if e.name == name {
			return e, true
		}
	}
	return SymbolEntity{}, false
}

func (s *SymbolTable) TypeOf(name string) string {
	for _, kind := range kinds {
		if e, ok := s.Contains(name, kind); ok {
			return e.typeName
		}
	}
	return ""
}

func (s *SymbolTable) IndexOf(name string) int {
	for _, kind := range kinds {
		if e, ok := s.Contains(name, kind); ok {
			return e.number
		}
	}
	return 0
}
