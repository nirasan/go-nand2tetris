package main

import (
	"bufio"
	"io"
	"strings"
)

var (
	commands = []string{
		`push`, `pop`, `add`, `sub`, `neg`, `eq`, `gt`, `lt`, `and`, `or`, `not`,
	}
	commandTypeMap = map[string]CommandType{
		`push`: C_PUSH,
		`pop`:  C_POP,
		`add`:  C_ARITHMETIC,
		`sub`:  C_ARITHMETIC,
		`neg`:  C_ARITHMETIC,
		`eq`:   C_ARITHMETIC,
		`gt`:   C_ARITHMETIC,
		`lt`:   C_ARITHMETIC,
		`and`:  C_ARITHMETIC,
		`or`:   C_ARITHMETIC,
		`not`:  C_ARITHMETIC,
	}
)

type CommandType uint8

const (
	C_NONE CommandType = iota
	C_ARITHMETIC
	C_PUSH
	C_POP
	C_LABEL
	C_GOTO
	C_IF
	C_FUNCTION
	C_RETURN
	C_CALL
)

type Parser struct {
	r    io.Reader
	s    *bufio.Scanner
	line string
}

func NewParder(r io.Reader) *Parser {
	return &Parser{
		r: r,
		s: bufio.NewScanner(r),
	}
}

func (p *Parser) HasMoreCommands() bool {
	if !p.s.Scan() {
		return false
	}

	line := strings.TrimSpace(p.s.Text())
	for _, command := range commands {
		if strings.Index(line, command) == 0 {
			p.line = line
			return true
		}
	}
	return p.HasMoreCommands()
}

func (p *Parser) Advance() string {
	return p.line
}

func (p *Parser) CommandType() CommandType {
	for k, v := range commandTypeMap {
		if strings.Index(p.line, k) == 0 {
			return v
		}
	}
	return C_NONE
}

func (p *Parser) Command() string {
	return strings.Split(p.line, ` `)[0]
}

func (p *Parser) Arg1() string {
	switch p.CommandType() {
	case C_RETURN:
		panic(`Invalid CommandType`)
	case C_ARITHMETIC:
		return p.Command()
	default:
		list := strings.Split(p.line, ` `)
		return list[1]
	}
}

func (p *Parser) Arg2() string {
	switch p.CommandType() {
	case C_PUSH, C_POP, C_FUNCTION, C_CALL:
		list := strings.Split(p.line, ` `)
		return list[2]
	default:
		panic(`Invalid CommandType`)
	}
}
