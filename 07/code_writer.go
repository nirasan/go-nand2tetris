package main

import (
	"io"
	"fmt"
)

type CodeWriter struct {
	w io.Writer
	filename string
}

func NewCodeWriter(w io.Writer) *CodeWriter {
	return &CodeWriter{w: w}
}

func (c *CodeWriter) SetFileName(f string) {
	c.filename = f
}

func (c *CodeWriter) WriteArithmetic(command string) {
	if command == "add" {
		c.p("@SP") // A = 0
		c.p("A=M") // A = M[0]
		c.p("A=A-1") // A = SP - 1
		c.p("D=M") // D = M[SP-1]
		c.p("A=A-1") // A = SP - 2
		c.p("M=D+M") // M[SP-2] = (SP-1) + M[SP-1]
		c.p("D=A+1") // D = SP - 1
		c.p("@SP") // A = 0
		c.p("M=D") // M[0] = SP - 1
	}
}

func (c *CodeWriter) WritePushPop(command string, segment string, index string) {
	if command == "push" {
		if segment == "constant" {
			c.p("@%s", index) // A = n
			c.p("D=A") // D = A
			c.p("@SP") // A = 0
			c.p("A=M") // A = M[0]
			c.p("M=D") // M[SP] = n
			c.p("@SP") // A = 0
			c.p("M=M+1") // M[0] = M[0] + 1
		}
	}
}

func (c *CodeWriter) p(format string, a ...interface{}) {
	fmt.Fprintf(c.w, format + "\n", a...)
}
