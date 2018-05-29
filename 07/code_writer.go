package main

import (
	"io"
	"fmt"
	"path"
	"strings"
)

var baseSymbolMap = map[string]string{
	"local": "LCL",
	"argument": "ARG",
	"this": "THIS",
	"that": "THAT",
	"pointer": "R3",
	"temp": "R5",
}

type CodeWriter struct {
	w io.Writer
	filename string
	lineNumber uint64
}

func NewCodeWriter(w io.Writer) *CodeWriter {
	return &CodeWriter{w: w}
}

func (c *CodeWriter) SetFileName(f string) {
	c.filename = strings.TrimRight(path.Base(f), ".vm")
}

func (c *CodeWriter) WriteArithmetic(command string) {
	c.l("// " + command)
	switch command {
	case "add":
		c.p("@SP") // A = 0
		c.p("A=M-1") // A = M[0] - 1
		c.p("D=M") // D = M[SP-1]
		c.p("A=A-1") // A = SP - 2
		c.p("M=D+M") // M[SP-2] = M[SP-1] + M[SP-2]
		c.p("D=A+1") // D = SP - 1
		c.p("@SP") // A = 0
		c.p("M=D") // M[0] = SP - 1
	case "sub":
		c.p("@SP") // A = 0
		c.p("A=M-1") // A = M[0] - 1
		c.p("D=M") // D = M[SP-1]
		c.p("A=A-1") // A = SP - 2
		c.p("M=M-D") // M[SP-2] = M[SP-2] - M[SP-1]
		c.p("D=A+1") // D = SP - 1
		c.p("@SP") // A = 0
		c.p("M=D") // M[0] = SP - 1
	case "and":
		c.p("@SP") // A = 0
		c.p("A=M-1") // A = M[0] - 1
		c.p("D=M") // D = M[SP-1]
		c.p("A=A-1") // A = SP - 2
		c.p("M=D&M") // M[SP-2] = M[SP-2] | M[SP-1]
		c.p("D=A+1") // D = SP - 1
		c.p("@SP") // A = 0
		c.p("M=D") // M[0] = SP - 1
	case "or":
		c.p("@SP") // A = 0
		c.p("A=M-1") // A = M[0] - 1
		c.p("D=M") // D = M[SP-1]
		c.p("A=A-1") // A = SP - 2
		c.p("M=D|M") // M[SP-2] = M[SP-2] & M[SP-1]
		c.p("D=A+1") // D = SP - 1
		c.p("@SP") // A = 0
		c.p("M=D") // M[0] = SP - 1
	case "neg":
		c.p("@SP") // A = 0
		c.p("A=M-1") // A = M[0] - 1
		c.p("M=-M") // M[SP-1] = -M[SP-1]
	case "not":
		c.p("@SP") // A = 0
		c.p("A=M-1") // A = M[0] - 1
		c.p("M=!M") // M[SP-1] = !M[SP-1]
	case "eq", "gt", "lt":
		c.p("@SP") // A = 0
		c.p("A=M-1") // A = M[0] - 1
		c.p("D=M") // D = M[SP-1]
		c.p("A=A-1") // A = SP - 2
		c.p("D=M-D") // D = M[SP-2] - M[SP-1]
		// compare
		c.p("@%d", c.lineNumber + 8) // A = true statement
		switch command {
		case "eq":
			c.p("D;JEQ") // if (D = 0) goto A
		case "gt":
			c.p("D;JGT") // if (D > 0) goto A
		case "lt":
			c.p("D;JLT") // if (D < 0) goto A
		}
		// set false
		c.p("@SP")
		c.p("A=M-1") // A = M[0] - 1 = SP - 1
		c.p("A=A-1") // A = SP - 2
		c.p("M=0") // M[SP-2] = false
		c.p("@%d", c.lineNumber + 6) // A = out of if statement
		c.p("0;JMP") // goto A
		// set true
		c.p("@SP")
		c.p("A=M-1") // A = M[0] - 1 = SP - 1
		c.p("A=A-1") // A = SP - 2
		c.p("M=-1") // M[SP-2] = true
		// set new SP
		c.p("@SP")
		c.p("M=M-1") // M[0] = M[0] - 1
	}
}

func (c *CodeWriter) WritePushPop(command string, segment string, index string) {
	c.l("// %s %s %s", command, segment, index)
	if command == "push" {
		switch segment {
		case "constant":
			c.p("@%s", index) // A = n
			c.p("D=A")        // D = A
			c.p("@SP")        // A = 0
			c.p("A=M")        // A = M[0]
			c.p("M=D")        // M[SP] = n
			c.p("@SP")        // A = 0
			c.p("M=M+1")      // M[0] = M[0] + 1
		case "local", "argument", "this", "that", "pointer", "temp":
			// get value
			c.p("@%s", index) // A = n
			c.p("D=A") // D = n
			switch segment {
			case "local", "argument", "this", "that":
				c.p("@%s", baseSymbolMap[segment]) // A = BASE address pointer
				c.p("A=M") // A = M[BASE address pointer] = BASE
			case "pointer", "temp":
				c.p("@%s", baseSymbolMap[segment]) // A = BASE
			}
			c.p("A=D+A") // A = BASE + n
			c.p("D=M") // D = M[BASE + n]
			// push
			c.p("@SP")
			c.p("A=M")
			c.p("M=D")
			// SP++
			c.p("@SP")
			c.p("M=M+1")
		case "static":
			// get value
			c.p("@%s.%s", c.filename, index) // A = FILENAME.INDEX
			c.p("D=M") // D = M[FILENAME.INDEX]
			// push
			c.p("@SP")
			c.p("A=M")
			c.p("M=D")
			// SP++
			c.p("@SP")
			c.p("M=M+1")
		}
	} else if command == "pop" {
		switch segment {
		case "local", "argument", "this", "that", "pointer", "temp":
			// calc local address
			c.p("@%s", index) // A = n
			c.p("D=A") // D = n
			switch segment {
			case "local", "argument", "this", "that":
				c.p("@%s", baseSymbolMap[segment]) // A = BASE address pointer
				c.p("A=M") // A = M[BASE address pointer] = BASE
			case "pointer", "temp":
				c.p("@%s", baseSymbolMap[segment]) // A = BASE
			}
			c.p("D=D+A") // D = BASE + n
			c.p("@R13") // A = 13 (common reg)
			c.p("M=D") // M[13] = BASE + n
			// get head of stack
			c.p("@SP") // A = 0
			c.p("A=M-1") // A = M[0] - 1
			c.p("D=M") // D = M[SP-1]
			// write
			c.p("@R13") // A = 13
			c.p("A=M") // A = BASE + n
			c.p("M=D") // M[BASE+n] = M[SP]
			// SP--
			c.p("@SP")
			c.p("M=M-1")
		case "static":
			// get head of stack
			c.p("@SP") // A = 0
			c.p("A=M-1") // A = M[0] - 1
			c.p("D=M") // D = M[SP-1]
			// write
			c.p("@%s.%s", c.filename, index) // A = FILENAME.INDEX
			c.p("M=D") // M[FILENAME.INDEX] = M[SP]
			// SP--
			c.p("@SP")
			c.p("M=M-1")
		}
	}
}

func (c *CodeWriter) p(format string, a ...interface{}) {
	fmt.Fprintf(c.w, format + "\n", a...)
	c.lineNumber += 1
}

func (c *CodeWriter) l(format string, a ...interface{}) {
	fmt.Fprintf(c.w, format + "\n", a...)
}
