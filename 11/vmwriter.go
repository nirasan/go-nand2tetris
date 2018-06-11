package main

import (
	"io"
	"fmt"
)

type VMWriter struct {
	w io.Writer
}

func NewVMWriter(w io.Writer) *VMWriter {
	return &VMWriter{w: w}
}

func (v *VMWriter) WritePush(segment string, index int) {
	fmt.Fprintf(v.w,"push %s %d\n", segment, index)
}

func (v *VMWriter) WritePop(segment string, index int) {
	fmt.Fprintf(v.w,"pop %s %d\n", segment, index)
}

func (v *VMWriter) WriteArithmetic(command string) {
	fmt.Fprintf(v.w,"%s\n", command)
}

func (v *VMWriter) WriteLabel(label string) {
	fmt.Fprintf(v.w,"label %s\n", label)
}

func (v *VMWriter) WriteGoto(label string) {
	fmt.Fprintf(v.w,"goto %s\n", label)
}

func (v *VMWriter) WriteIf(label string) {
	fmt.Fprintf(v.w,"if-goto %s\n", label)
}

func (v *VMWriter) WriteCall(name string, number int) {
	fmt.Fprintf(v.w,"call %s %d\n", name, number)
}

func (v *VMWriter) WriteFunction(name string, number int) {
	fmt.Fprintf(v.w,"function %s %d\n", name, number)
}

func (v *VMWriter) WriteReturn() {
	fmt.Fprint(v.w,"return\n")
}

