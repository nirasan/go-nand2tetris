package main

import (
	"bufio"
	"io"
	"log"
	"strings"
	"strconv"
)

var (
	symbols = []byte{
		'{', '}', '(', ')', '[', ']', '.', ',', ';', '+', '-', '*', '/', '&', '|', '<', '>', '=', '~',
	}
	keywords = []string{
		"class", "constructor", "function", "method", "field", "static", "var", "int", "char", "boolean",
		"void", "true", "false", "null", "this", "let", "do", "if", "else", "while", "return",
	}
)

type TokenType uint8

const (
	TOKEN_NONE TokenType = iota
	TOKEN_KEYWORD
	TOKEN_SYMBOL
	TOKEN_IDENTIFIER
	TOKEN_INT_CONST
	TOKEN_STRING_CONST
)

type Tokenizer struct {
	r         io.Reader
	s         *bufio.Scanner
	line      string
	index     int
	token     string
	tokenType TokenType
}

func NewTokenizer(r io.Reader) *Tokenizer {
	return &Tokenizer{r: r, s: bufio.NewScanner(r)}
}

func (t *Tokenizer) HasMoreTokens() bool {
	log.Printf("%#v", t)
	if t.line == "" || len(t.line) <= t.index {
		log.Printf(t.line)
		if !t.s.Scan() {
			return false
		}

		line := strings.TrimSpace(t.s.Text())

		if len(line) == 0 {
			return t.HasMoreTokens()
		}

		t.line = line
		t.index = 0
	}
	log.Printf(t.line)
	c := t.line[t.index]
	l := string(t.line[t.index:])

	// skip space
	if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
		t.index += 1
		return t.HasMoreTokens()
	}

	// skip '//' comment
	if strings.Index(l, "//") == 0 {
		t.index = len(t.line)
		return t.HasMoreTokens()
	}

	// skip '/* */' comment
	if strings.Index(l, "/*") == 0 {
		var i int
		for {
			i = strings.Index(l, "*/")
			if i != -1 {
				break
			}
			// read next line
			if !t.s.Scan() {
				return false
			}
			line := strings.TrimSpace(t.s.Text())
			t.line = line
			t.index = 0
			l = t.line
		}
		t.index = t.index + i + 2
		return t.HasMoreTokens()
	}

	// symbol?
	if IsSymbol(c) {
		t.token = string([]byte{c})
		t.tokenType = TOKEN_SYMBOL
		t.index += 1
		return true
	}

	// number?
	if IsNumber(c) {
		list := []byte{c}
		i := t.index + 1
		for i < len(t.line) {
			if IsNumber(t.line[i]) {
				list = append(list, t.line[i])
				i += 1
			} else {
				break
			}
		}
		t.token = string(list)
		t.tokenType = TOKEN_INT_CONST
		t.index = i
		return true
	}

	// string?
	if c == '"' {
		i := strings.Index(string(t.line[t.index+1:]), `"`)
		if i == -1 {
			panic("string literal close word not found: " + t.line)
		}
		t.token = string(t.line[t.index+1 : t.index+i+1])
		t.tokenType = TOKEN_STRING_CONST
		t.index = t.index + i + 2
		return true
	}

	// keyword?

	// identifier?
	if IsIdentifierHead(c) {
		list := []byte{c}
		i := t.index + 1
		for i < len(t.line) {
			if IsIdentifier(t.line[i]) {
				list = append(list, t.line[i])
				i += 1
			} else {
				break
			}
		}
		t.token = string(list)
		t.index = i
		if IsKeyword(t.token) {
			t.tokenType = TOKEN_KEYWORD
		} else {
			t.tokenType = TOKEN_IDENTIFIER
		}
		return true
	}

	return false
}

func (t *Tokenizer) TokenType() TokenType {
	return t.tokenType
}

func (t *Tokenizer) KeyWord() string {
	if t.tokenType == TOKEN_KEYWORD {
		return t.token
	}
	return ""
}

func (t *Tokenizer) Symbol() string {
	if t.tokenType == TOKEN_SYMBOL {
		switch t.token {
		case "<":
			return "&lt;"
		case ">":
			return "&gt;"
		case "&":
			return "&amp;"
		default:
			return t.token
		}
	}
	return ""
}

func (t *Tokenizer) Identifier() string {
	if t.tokenType == TOKEN_IDENTIFIER {
		return t.token
	}
	return ""
}

func (t *Tokenizer) IntVal() int {
	if t.tokenType == TOKEN_INT_CONST {
		i, err := strconv.Atoi(t.token)
		if err != nil {
			panic(err)
		}
		return i
	}
	return 0
}

func (t *Tokenizer) StringVal() string {
	if t.tokenType == TOKEN_STRING_CONST {
		return t.token
	}
	return ""
}

func IsAlpha(b byte) bool {
	return ('A' <= b && b <= 'Z') || ('a' <= b && b <= 'z')
}

func IsNumber(b byte) bool {
	return '0' <= b && b <= '9'
}

func IsSymbol(b byte) bool {
	for _, s := range symbols {
		if b == s {
			return true
		}
	}
	return false
}

func IsIdentifierHead(b byte) bool {
	return IsAlpha(b) || b == '_'
}

func IsIdentifier(b byte) bool {
	return IsIdentifierHead(b) || IsNumber(b)
}

func IsKeyword(s string) bool {
	for _, k := range keywords {
		if s == k {
			return true
		}
	}
	return false
}
