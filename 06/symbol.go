package main

type SymbolTable struct {
	table map[string]int
}

func NewSymbolTable() *SymbolTable {
	s := &SymbolTable{
		make(map[string]int),
	}
	s.AddEntry("SP", 0)
	s.AddEntry("LCL", 1)
	s.AddEntry("ARG", 2)
	s.AddEntry("THIS", 3)
	s.AddEntry("THAT", 4)
	s.AddEntry("R0", 0)
	s.AddEntry("R1", 1)
	s.AddEntry("R2", 2)
	s.AddEntry("R3", 3)
	s.AddEntry("R4", 4)
	s.AddEntry("R5", 5)
	s.AddEntry("R6", 6)
	s.AddEntry("R7", 7)
	s.AddEntry("R8", 8)
	s.AddEntry("R9", 9)
	s.AddEntry("R10", 10)
	s.AddEntry("R11", 11)
	s.AddEntry("R12", 12)
	s.AddEntry("R13", 13)
	s.AddEntry("R14", 14)
	s.AddEntry("R15", 15)
	s.AddEntry("SCREEN", 16384)
	s.AddEntry("KBD", 24576)
	return s
}

func (s *SymbolTable) AddEntry(symbol string, addr int) {
	s.table[symbol] = addr
}

func (s *SymbolTable) Contains(symbol string) bool {
	_, ok := s.table[symbol]
	return ok
}

func (s *SymbolTable) GetAddress(symbol string) int {
	return s.table[symbol]
}
