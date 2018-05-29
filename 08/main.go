package main

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	filename := os.Args[1]
	stat, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}

	codeWriter := NewCodeWriter(os.Stdout)

	var files []*os.File
	if stat.IsDir() {
		list, err := ioutil.ReadDir(filename)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range list {
			if !strings.HasSuffix(f.Name(), ".vm") {
				continue
			}
			ff, err := os.Open(filename + f.Name())
			if err != nil {
				log.Fatal(err)
			}
			files = append(files, ff)
		}
		codeWriter.WriteInit()
	} else {
		f, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		files = append(files, f)
	}

	for _, f := range files {
		log.Printf("FILE: %s", f.Name())
		codeWriter.SetFileName(f.Name())
		parser := NewParder(f)

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
}
