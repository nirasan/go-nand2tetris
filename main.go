package main

import (
	"fmt"
	"log"
	"os"

	"strconv"
)

func main() {
	filename := os.Args[1]
	parser, err := NewParser(filename)
	if err != nil {
		log.Fatal(err)
	}
	code := NewCode()
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
			out, err := strconv.ParseInt(parser.Symbol(), 10, 16)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%016b\n", uint16(out))
		}
	}
}
