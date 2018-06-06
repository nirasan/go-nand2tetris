package main

import (
	"io"
	"bufio"
	"fmt"
	"regexp"
	"log"
	"strings"
)

var (
	tagIdentifier = "identifier"
	regexTag = regexp.MustCompile(`^<(.*?)> (.*) </.*>$`)
)

type Compiler struct {
	r io.Reader
	w io.Writer
	s *bufio.Scanner
	line string
	tag string
	value string
	indent int
}

func NewCompiler(r io.Reader, w io.Writer) *Compiler {
	return &Compiler{
		r: r,
		w: w,
		s: bufio.NewScanner(r),
	}
}

func (c *Compiler) CompileClass() {
	// <tokens>
	c.Scan()
	c.LineMust("<tokens>")
	c.Println("<class>")
	c.indent += 1

	// class
	c.Scan()
	c.ValueMust("class")
	c.PrintLine()

	// className
	c.Scan()
	c.TagMust(tagIdentifier)
	c.PrintLine()

	// {
	c.Scan()
	c.ValueMust("{")
	c.PrintLine()

	// classVarDec*
	c.Scan()
	for c.TagIs("keyword") && c.ValueIs("static", "field") {
		c.CompileClassVarDec()
	}

	// subroutineDec*
	for c.TagIs("keyword") && c.ValueIs("constructor", "function", "method") {
		c.CompileSubroutineDec()
	}

	// }
	c.ValueMust("}")
	c.PrintLine()

	// close class
	c.indent -= 1
	c.Println("</class>")
}

func (c *Compiler) CompileClassVarDec() {
	// classVarDec
	c.Println("<classVarDec>")
	c.indent += 1

	// static | field
	c.TagMust("keyword")
	c.ValueMust("static", "field")
	c.PrintLine()

	// type
	c.Scan()
	c.MustType()
	c.PrintLine()
	
	// varName
	c.Scan()
	c.TagMust(tagIdentifier)
	c.PrintLine()
	
	// (, varName)*
	c.Scan()
	for c.ValueIs(",") {
		// ,
		c.PrintLine()
		c.Scan()
		// varName
		c.TagMust(tagIdentifier)
		c.PrintLine()
		c.Scan()
	}
	
	// ;
	c.ValueMust(";")
	c.PrintLine()
	c.Scan()

	// close classVarDec
	c.indent -= 1
	c.Println("</classVarDec>")
}

func (c *Compiler) CompileSubroutineDec() {
	// subroutineDec
	c.Println("<subroutineDec>")
	c.indent += 1

	// constructor, function, method
	c.ValueMust("constructor", "function", "method")
	c.PrintLine()
	c.Scan()
	
	// void | type
	if !c.ValueIs("void") && !c.IsType() {
		log.Fatalf("invalid type: %#v", c)
	}
	c.PrintLine()
	c.Scan()

	// subroutineName
	c.TagMust(tagIdentifier)
	c.PrintLine()
	c.Scan()

	// (
	c.ValueMust("(")
	c.PrintLine()
	c.Scan()

	// parameterList
	c.CompileParameterList()

	// )
	c.ValueMust(")")
	c.PrintLine()
	c.Scan()

	// subroutineBody
	c.CompileSubroutineBody()

	// close subroutineDec
	c.indent -= 1
	c.Println("</subroutineDec>")
}

// ( (type varName) ((, type varName))* )?
func (c *Compiler) CompileParameterList() {
	// parameterList
	c.Println("<parameterList>")
	c.indent += 1

	// empty parameterList
	if c.IsType() {

		// type
		c.MustType()
		c.PrintLine()
		c.Scan()

		// varName
		c.TagMust(tagIdentifier)
		c.PrintLine()
		c.Scan()

		// (, type varName)*
		for c.ValueIs(",") {
			// ,
			c.PrintLine()
			c.Scan()

			// type
			c.MustType()
			c.PrintLine()
			c.Scan()

			// varName
			c.TagMust(tagIdentifier)
			c.PrintLine()
			c.Scan()
		}
	}

	// close parameterList
	c.indent -=1
	c.Println("</parameterList>")
}

// { varDec* statements }
func (c *Compiler) CompileSubroutineBody() {
	// subroutineBody
	c.Println("<subroutineBody>")
	c.indent += 1

	// {
	c.ValueMust("{")
	c.PrintLine()
	c.Scan()

	// varDec*
	for c.ValueIs("var") {
		c.CompileVarDec()
	}

	// statements
	c.CompileStatements()

	// }
	c.ValueMust("}")
	c.PrintLine()
	c.Scan()

	// close subroutineBody
	c.indent -=1
	c.Println("</subroutineBody>")
}

// var type varName (, varName)* ;
func (c *Compiler) CompileVarDec() {
	// varDec
	c.Println("<varDec>")
	c.indent += 1

	// var
	c.ValueMust("var")
	c.PrintLine()
	c.Scan()

	// type
	c.MustType()
	c.PrintLine()
	c.Scan()

	// varName
	c.TagMust(tagIdentifier)
	c.PrintLine()
	c.Scan()

	// (, varName)*
	for c.ValueIs(",") {
		// ,
		c.PrintLine()
		c.Scan()
		// varName
		c.TagMust(tagIdentifier)
		c.PrintLine()
		c.Scan()
	}

	// ;
	c.ValueMust(";")
	c.PrintLine()
	c.Scan()

	// close varDec
	c.indent -= 1
	c.Println("</varDec>")
}

func (c *Compiler) CompileStatements() {
	// statements
	c.Println("<statements>")
	c.indent += 1

	// statement*
	for c.ValueIs("let", "if", "while", "do", "return") {
		switch c.Value() {
		case "let":
			c.CompileLetStatement()
		case "if":
			c.CompileIfStatement()
		case "while":
			c.CompileWhileStatement()
		case "do":
			c.CompileDoStatement()
		case "return":
			c.CompileReturnStatement()
		}
	}

	// close statements
	c.indent -= 1
	c.Println("</statements>")
}

func (c *Compiler) CompileLetStatement() {
	// letStatement
	c.Println("<letStatement>")
	c.indent += 1

	// let
	c.ValueMust("let")
	c.PrintLine()

	// varName
	c.Scan()
	c.TagMust(tagIdentifier)
	c.PrintLine()

	// ( [ expression ] )?
	c.Scan()
	if c.ValueIs("[") {
		// [
		c.PrintLine()
		c.Scan()

		// expression
		c.CompileExpression()

		// ]
		c.ValueMust("]")
		c.PrintLine()
		c.Scan()
	}

	// =
	c.ValueMust("=")
	c.PrintLine()
	c.Scan()

	// expression
	c.CompileExpression()

	// ;
	c.ValueMust(";")
	c.PrintLine()
	c.Scan()

	// close letStatement
	c.indent -= 1
	c.Println("</letStatement>")
}

func (c *Compiler) CompileReturnStatement() {
	// returnStatement
	c.Println("<returnStatement>")
	c.indent += 1

	// return
	c.ValueMust("return")
	c.PrintLine()
	c.Scan()

	// expression?
	if !c.ValueIs(";") {
		c.CompileExpression()
	}

	// ;
	c.ValueMust(";")
	c.PrintLine()
	c.Scan()

	// close returnStatement
	c.indent -= 1
	c.Println("</returnStatement>")
}

func (c *Compiler) CompileIfStatement() {
	// ifStatement
	c.Println("<ifStatement>")
	c.indent += 1

	// if
	c.ValueMust("if")
	c.PrintLine()
	c.Scan()

	// (
	c.ValueMust("(")
	c.PrintLine()
	c.Scan()

	// expression
	c.CompileExpression()

	// )
	c.ValueMust(")")
	c.PrintLine()
	c.Scan()

	// {
	c.ValueMust("{")
	c.PrintLine()
	c.Scan()

	// statements
	c.CompileStatements()

	// }
	c.ValueMust("}")
	c.PrintLine()
	c.Scan()

	// ( else { statements } )?
	if c.ValueIs("else") {
		// else
		c.PrintLine()
		c.Scan()

		// {
		c.ValueMust("{")
		c.PrintLine()
		c.Scan()

		// statements
		c.CompileStatements()

		// }
		c.ValueMust("}")
		c.PrintLine()
		c.Scan()
	}

	// close ifStatement
	c.indent -= 1
	c.Println("</ifStatement>")
}

func (c *Compiler) CompileWhileStatement() {
	// whileStatement
	c.Println("<whileStatement>")
	c.indent += 1

	// while
	c.ValueMust("while")
	c.PrintLine()
	c.Scan()

	// (
	c.ValueMust("(")
	c.PrintLine()
	c.Scan()

	// expression
	c.CompileExpression()

	// )
	c.ValueMust(")")
	c.PrintLine()
	c.Scan()

	// {
	c.ValueMust("{")
	c.PrintLine()
	c.Scan()

	// statements
	c.CompileStatements()

	// }
	c.ValueMust("}")
	c.PrintLine()
	c.Scan()

	// close whileStatement
	c.indent -= 1
	c.Println("</whileStatement>")
}

func (c *Compiler) CompileDoStatement() {
	// doStatement
	c.Println("<doStatement>")
	c.indent += 1

	// do
	c.ValueMust("do")
	c.PrintLine()
	c.Scan()

	// subroutineCall
	c.CompileSubroutineCall()

	// ;
	c.ValueMust(";")
	c.PrintLine()
	c.Scan()

	// close doStatement
	c.indent -= 1
	c.Println("</doStatement>")
}


func (c *Compiler) CompileExpression() {
	// expression
	c.Println("<expression>")
	c.indent += 1

	// term
	c.CompileTerm()

	// (op term)*
	for c.IsOp() {
		// op
		c.PrintLine()
		c.Scan()
		// term
		c.CompileTerm()
	}

	// close expression
	c.indent -= 1
	c.Println("</expression>")
}

func (c *Compiler) CompileTerm() {
	// term
	c.Println("<term>")
	c.indent += 1

	if c.TagIs("integerConstant") {
		// integerConstant
		c.PrintLine()
		c.Scan()
	} else if c.TagIs("stringConstant") {
		// stringConstant
		c.PrintLine()
		c.Scan()
	} else if c.TagIs("keyword") && c.ValueIs("true", "false", "null", "this") {
		// true | false | null | this
		c.PrintLine()
		c.Scan()
	} else if c.ValueIs("(") {
		// (
		c.PrintLine()
		c.Scan()

		// expression
		c.CompileExpression()

		// )
		c.ValueMust(")")
		c.PrintLine()
		c.Scan()
	} else if c.TagIs("symbol") && c.ValueIs("-", "~") {
		// - | ~
		c.PrintLine()
		c.Scan()
		// term
		c.CompileTerm()
	} else if c.TagIs(tagIdentifier) {
		// subroutineCall
		c.CompileSubroutineCall()
	}

	// close term
	c.indent -= 1
	c.Println("</term>")
}

// subroutineCall
func (c *Compiler) CompileSubroutineCall() {
	// varName | varName [ expression ] | subroutineName ( expressionList ) | (className | varName).subroutineName ( expressionList )
	c.TagMust(tagIdentifier)
	c.PrintLine()
	c.Scan()

	if c.ValueIs("[") {
		// varName [ expression ]

		// [
		c.PrintLine()
		c.Scan()

		// expression
		c.CompileExpression()

		// ]
		c.ValueMust("]")
		c.PrintLine()
		c.Scan()

	} else if c.ValueIs(".") {
		// (className | varName) . subroutineName ( expressionList )

		// .
		c.PrintLine()
		c.Scan()

		// subroutineName
		c.TagMust(tagIdentifier)
		c.PrintLine()
		c.Scan()

		// (
		c.ValueMust("(")
		c.PrintLine()
		c.Scan()

		// expressionList
		c.CompileExpressionList()

		// )
		c.ValueMust(")")
		c.PrintLine()
		c.Scan()

	} else if c.ValueIs("(") {
		// subroutineName ( expressionList )

		// (
		c.ValueMust("(")
		c.PrintLine()
		c.Scan()

		// expressionList
		c.CompileExpressionList()

		// )
		c.ValueMust(")")
		c.PrintLine()
		c.Scan()

	} else {
		// varName
		// no op
	}
}

// ( expression (, expression)* )?
func (c *Compiler) CompileExpressionList() {
	// expressionList
	c.Println("<expressionList>")
	c.indent += 1

	if c.IsTerm() {

		// expression
		c.CompileExpression()

		for c.ValueIs(",") {
			// ,
			c.PrintLine()
			c.Scan()

			// expression
			c.CompileExpression()
		}

	}

	// close expressionList
	c.indent -= 1
	c.Println("</expressionList>")
}

func (c *Compiler) Scan() bool {
	if c.s.Scan() {
		c.line = strings.TrimSpace(c.s.Text())
		c.tag = ""
		c.value = ""
		log.Printf("SCAN: %s\n", c.line)
		return true
	} else {
		return false
	}
}

func (c *Compiler) Tag() string {
	tag, _ := c.ParseLine()
	return tag
}

func (c *Compiler) Value() string {
	_, value := c.ParseLine()
	return value
}

func (c *Compiler) ParseLine() (string, string) {
	if c.tag == "" && c.value == "" {
		m := regexTag.FindAllStringSubmatch(c.line, -1)
		if len(m) != 1 {
			log.Fatalf("invalid match result: %#v", m)
		}
		c.tag, c.value = m[0][1], m[0][2]
		switch c.value {
		case "&lt;":
			c.value = "<"
		case "&gt;":
			c.value = ">"
		case "&amp;":
			c.value = "&"
		}
	}
	return c.tag, c.value
}

func (c *Compiler) TagMust(tags ...string) {
	if !c.TagIs(tags...) {
		log.Fatalf("TagMust %v, but %s. %#v", tags, c.Tag(), c)
	}
}

func (c *Compiler) ValueMust(values ...string) {
	if !c.ValueIs(values...) {
		log.Fatalf("ValueMust %v, but %s. %#v", values, c.Value(), c)
	}
}

func (c *Compiler) LineMust(s string) {
	if c.line != s {
		log.Fatalf("LineMust %s, but %s. %#v", s, c.line, c)
	}
}

func (c *Compiler) TagIs(tags ...string) bool {
	match := false
	for _, t := range tags {
		if c.Tag() == t {
			match = true
			break
		}
	}
	return match
}

func (c *Compiler) ValueIs(values ...string) bool {
	match := false
	for _, v := range values {
		if c.Value() == v {
			match = true
			break
		}
	}
	return match
}

func (c *Compiler) Println(s string) {
	for i := 0; i < c.indent; i++ {
		fmt.Fprint(c.w, "  ")
	}
	fmt.Fprintln(c.w, s)
}

func (c *Compiler) PrintLine() {
	c.Println(c.line)
}

func (c *Compiler) IsType() bool {
	return (c.TagIs("keyword") && c.ValueIs("int", "char", "boolean")) || c.TagIs("identifier")
}

func (c *Compiler) MustType() {
	if !c.IsType() {
		log.Fatalf("token must type: %#v", c)
	}
}

func (c *Compiler) IsOp() bool {
	return c.ValueIs("+", "-", "*", "/", "&", "|", "<", ">", "=")
}

func (c *Compiler) IsTerm() bool {
	if c.TagIs("integerConstant") {
		return true
	} else if c.TagIs("stringConstant") {
		return true
	} else if c.TagIs("keyword") && c.ValueIs("true", "false", "null", "this") {
		return true
	} else if c.ValueIs("(") {
		return true
	} else if c.TagIs("symbol") && c.ValueIs("-", "~") {
		return true
	} else if c.TagIs(tagIdentifier) {
		return true
	}
	return false
}
