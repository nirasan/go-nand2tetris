package main

import (
	"os"
	"log"
)

func main() {
	filename := os.Args[1]
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	parser := NewParder(f)
	codeWriter := NewCodeWriter(os.Stdout)
	codeWriter.SetFileName(filename)

	for parser.HasMoreCommands() {
		log.Printf("%s", parser.line)
		switch parser.CommandType() {
		case C_ARITHMETIC:
			codeWriter.WriteArithmetic(parser.Arg1())
		case C_PUSH, C_POP:
			codeWriter.WritePushPop(parser.Command(), parser.Arg1(), parser.Arg2())
		}
	}
}
