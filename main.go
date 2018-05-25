package main

import (
	"fmt"
	"log"
	"os"

	"regexp"
	"strconv"
)

var numberRegex = regexp.MustCompile(`^\d+$`)

func main() {
	filename := os.Args[1]
	parser, err := NewParser(filename)
	if err != nil {
		log.Fatal(err)
	}
	// Add label to symbolTable
	symbolTable := NewSymbolTable()
	romAddr := 0
	for parser.HasMoreCommands() {
		if parser.CommandType() == L_COMMAND {
			symbolTable.AddEntry(parser.Symbol(), romAddr)
		} else {
			romAddr += 1
		}
	}
	// Output binary
	parser, err = NewParser(filename)
	if err != nil {
		log.Fatal(err)
	}
	code := NewCode()
	ramAddr := 16
	for parser.HasMoreCommands() {
		if parser.CommandType() == C_COMMAND {
			var out uint16 = (1 << 15) | (1 << 14) | (1 << 13)
			comp := code.Comp(parser.Comp())
			dest := code.Dest(parser.Dest())
			jump := code.Jump(parser.Jump())
			out = out | (comp << 6) | (dest << 3) | jump
			log.Printf("Line: %s, Out: %016b, Comp: %b, Dest: %b, Jump: %b", parser.CurrentLine, out, comp, dest, jump)
			fmt.Printf("%016b\n", out)
		} else if parser.CommandType() == A_COMMAND {
			var addr int64 = 0
			symbol := parser.Symbol()
			if numberRegex.MatchString(symbol) {
				addr, err = strconv.ParseInt(symbol, 10, 16)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				if symbolTable.Contains(symbol) {
					addr = int64(symbolTable.GetAddress(symbol))
				} else {
					symbolTable.AddEntry(symbol, ramAddr)
					addr = int64(ramAddr)
					ramAddr += 1
				}
			}
			log.Printf("Line: %s, Out: %016b, Symbol: %s, Addr: %d", parser.CurrentLine, uint16(addr), symbol, addr)
			fmt.Printf("%016b\n", uint16(addr))
		}
	}
}
