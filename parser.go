package main

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

var (
	ACommandRegex = regexp.MustCompile(`^\s*@(.*)`)
	CCommandRegex = regexp.MustCompile(`^\s*([ADM]+=)?([\-\+\!\&\|01ADM]+)(;[JGELNMTQP]+)?`)
	LCommandRegex = regexp.MustCompile(`^\s*\((.*)\)`)
)

type CommandType uint8

const (
	INVALID_COMMAND CommandType = iota
	A_COMMAND
	C_COMMAND
	L_COMMAND
)

type Parser struct {
	Scanner     *bufio.Scanner
	CurrentLine string
}

func NewParser(filename string) (*Parser, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(f)
	return &Parser{
		Scanner: scanner,
	}, nil
}

func (p *Parser) HasMoreCommands() bool {
	hasMoreLine := p.Scanner.Scan()
	if !hasMoreLine {
		return false
	}
	line := p.Scanner.Text()
	if ACommandRegex.MatchString(line) || CCommandRegex.MatchString(line) || LCommandRegex.MatchString(line) {
		p.CurrentLine = line
		return true
	}
	return p.HasMoreCommands()
}

func (p *Parser) Advance() string {
	return p.CurrentLine
}

func (p *Parser) CommandType() CommandType {
	if ACommandRegex.MatchString(p.CurrentLine) {
		return A_COMMAND
	}
	if CCommandRegex.MatchString(p.CurrentLine) {
		return C_COMMAND
	}
	if LCommandRegex.MatchString(p.CurrentLine) {
		return L_COMMAND
	}
	return INVALID_COMMAND
}

func (p *Parser) Symbol() string {
	if p.CommandType() == A_COMMAND {
		m := ACommandRegex.FindAllStringSubmatch(p.CurrentLine, -1)
		return m[0][1]
	}
	if p.CommandType() == L_COMMAND {
		m := LCommandRegex.FindAllStringSubmatch(p.CurrentLine, -1)
		return m[0][1]
	}
	return ""
}

func (p *Parser) Dest() string {
	return strings.TrimSuffix(p.cCommandSubstring(1), `=`)
}

func (p *Parser) Comp() string {
	return p.cCommandSubstring(2)
}

func (p *Parser) Jump() string {
	return strings.TrimPrefix(p.cCommandSubstring(3), `;`)

}

func (p *Parser) cCommandSubstring(n int) string {
	if p.CommandType() == C_COMMAND {
		m := CCommandRegex.FindAllStringSubmatch(p.CurrentLine, -1)
		return m[0][n]
	}
	return ""
}
