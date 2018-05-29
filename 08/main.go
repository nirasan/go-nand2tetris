package main

import (
	"log"
	"os"
	"strconv"
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
		case C_LABEL:
			codeWriter.WriteLabel(parser.Arg1())
		case C_GOTO:
			codeWriter.WriteGoto(parser.Arg1())
		case C_IF:
			codeWriter.WriteIf(parser.Arg1())
		case C_CALL:
			i, err := strconv.Atoi(parser.Arg2())
			if err != nil {
				panic(err)
			}
			codeWriter.WriteCall(parser.Arg1(), i)
		case C_RETURN:
			codeWriter.WriteReturn()
		case C_FUNCTION:
			i, err := strconv.Atoi(parser.Arg2())
			if err != nil {
				panic(err)
			}
			codeWriter.WriteFunction(parser.Arg1(), i)
		}
	}
}
