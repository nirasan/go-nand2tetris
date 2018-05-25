package main

import "strings"

type Code struct {
}

func NewCode() *Code {
	return &Code{}
}

const (
	destMBit uint16 = 1 << iota
	destDBit
	destABit
)

const (
	jumpNull uint16 = iota
	jumpJGT
	jumpJEQ
	jumpJGE
	jumpJLT
	jumpJNE
	jumpJLE
	jumpJMP
)

const (
	compC6 uint16 = 1 << iota
	compC5
	compC4
	compC3
	compC2
	compC1
	compA
)

func (c *Code) Dest(s string) uint16 {
	var out uint16
	if strings.Index(s, `M`) > -1 {
		out = out | destMBit
	}
	if strings.Index(s, `D`) > -1 {
		out = out | destDBit
	}
	if strings.Index(s, `A`) > -1 {
		out = out | destABit
	}
	return out
}

func (c *Code) Jump(s string) uint16 {
	var out uint16
	switch s {
	case `JGT`:
		out = jumpJGT
	case `JEQ`:
		out = jumpJEQ
	case `JGE`:
		out = jumpJGE
	case `JLT`:
		out = jumpJLT
	case `JNE`:
		out = jumpJNE
	case `JLE`:
		out = jumpJLE
	case `JMP`:
		out = jumpJMP
	}
	return out
}

func (c *Code) Comp(s string) uint16 {
	var out uint16
	if strings.Index(s, `M`) > -1 {
		out = out | compA
	}
	switch s {
	case `0`:
		out = out | compC1 | compC3 | compC5
	case `1`:
		out = out | compC1 | compC2 | compC3 | compC4 | compC5 | compC6
	case `-1`:
		out = out | compC1 | compC2 | compC3 | compC5
	case `D`:
		out = out | compC3 | compC4
	case `A`, `M`:
		out = out | compC1 | compC2
	case `!D`:
		out = out | compC3 | compC4 | compC6
	case `!A`, `!M`:
		out = out | compC1 | compC2 | compC6
	case `-D`:
		out = out | compC3 | compC4 | compC5 | compC6
	case `-A`, `-M`:
		out = out | compC1 | compC2 | compC5 | compC6
	case `D+1`:
		out = out | compC2 | compC3 | compC4 | compC5 | compC6
	case `A+1`, `M+1`:
		out = out | compC1 | compC2 | compC4 | compC5 | compC6
	case `D-1`:
		out = out | compC3 | compC4 | compC5
	case `A-1`, `M-1`:
		out = out | compC1 | compC2 | compC5
	case `D+A`, `D+M`:
		out = out | compC5
	case `D-A`, `D-M`:
		out = out | compC2 | compC5 | compC6
	case `A-D`, `M-D`:
		out = out | compC4 | compC5 | compC6
	case `D&A`, `D&M`:
		out = out
	case `D|A`, `D|M`:
		out = out | compC2 | compC4 | compC6
	}
	return out
}
