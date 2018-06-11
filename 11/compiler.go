package main

import (
	"io"
	"bufio"
	"fmt"
	"regexp"
	"log"
	"strings"
	"strconv"
	"os"
)

var (
	tagIdentifier = "identifier"
	regexTag = regexp.MustCompile(`^<(.*?)> (.*) </.*>$`)
	kindToSegment = map[string]string{
		"field": "this",
		"static": "static",
		"local": "local",
		"arg": "argument",
	}
)

type Compiler struct {
	r io.Reader
	w io.Writer
	s *bufio.Scanner
	line string
	tag string
	value string
	indent int

	symbolTable *SymbolTable
	vmWriter *VMWriter

	className string
	subroutineName string
	subroutineKind string
	expressionListCount int
	labelCount int
}

func NewCompiler(r io.Reader, w io.Writer, ww io.Writer) *Compiler {
	c := &Compiler{
		r: r,
		w: w,
		s: bufio.NewScanner(r),
		symbolTable: NewSymbolTable(),
		vmWriter: NewVMWriter(ww),
	}

	// create subroutine symbol table
	file := r.(*os.File)
	s := bufio.NewScanner(file)
	for s.Scan() {
		line := s.Text()
		if strings.Index(line, "<keyword>") > -1 {
			// subroutine kind
			kind := regexTag.FindAllStringSubmatch(line, -1)[0][2]
			if kind == "method" || kind == "constructor" || kind == "function" {
				// subroutine return type
				s.Scan()
				line = s.Text()
				typeName := regexTag.FindAllStringSubmatch(line, -1)[0][2]
				// subroutine name
				s.Scan()
				line = s.Text()
				name := regexTag.FindAllStringSubmatch(line, -1)[0][2]
				// set symbol table
				c.symbolTable.Define(name, typeName, kind)
			}
		}
	}
	file.Seek(0, 0)

	return c
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
	c.className = c.value
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
	kind := c.value

	// type
	c.Scan()
	c.MustType()
	c.PrintLine()
	typeName := c.value
	
	// varName
	c.Scan()
	c.TagMust(tagIdentifier)
	c.PrintLine()
	name := c.value

	c.symbolTable.Define(name, typeName, kind)
	
	// (, varName)*
	c.Scan()
	for c.ValueIs(",") {
		// ,
		c.PrintLine()
		c.Scan()
		// varName
		c.TagMust(tagIdentifier)
		c.PrintLine()
		name = c.value
		c.Scan()

		c.symbolTable.Define(name, typeName, kind)
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

	c.symbolTable.StartSubroutine()

	// constructor, function, method
	c.ValueMust("constructor", "function", "method")
	c.PrintLine()
	c.subroutineKind = c.value
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
	c.subroutineName = c.value
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

	if c.subroutineKind == "method" {
		c.symbolTable.Define("this", c.className, "arg")
	}

	// empty parameterList
	if c.IsType() {

		// type
		c.MustType()
		c.PrintLine()
		typeName := c.value
		c.Scan()

		// varName
		c.TagMust(tagIdentifier)
		c.PrintLine()
		name := c.value
		c.Scan()

		c.symbolTable.Define(name, typeName, "arg")

		// (, type varName)*
		for c.ValueIs(",") {
			// ,
			c.PrintLine()
			c.Scan()

			// type
			c.MustType()
			c.PrintLine()
			typeName = c.value
			c.Scan()

			// varName
			c.TagMust(tagIdentifier)
			c.PrintLine()
			name = c.value
			c.symbolTable.Define(name, typeName, "arg")
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

	c.vmWriter.WriteFunction(fmt.Sprintf("%s.%s", c.className, c.subroutineName), c.symbolTable.VarCount("local"))

	if c.subroutineKind == "method" {
		// read this
		c.vmWriter.WritePush("argument", 0)
		// set this
		c.vmWriter.WritePop("pointer", 0)
	} else if c.subroutineKind == "constructor" {
		// get field size
		c.vmWriter.WritePush("constant", c.symbolTable.VarCount("field"))
		// call alloc memory
		c.vmWriter.WriteCall("Memory.alloc", 1)
		// get this address
		c.vmWriter.WritePop("pointer", 0)
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
	typeName := c.value
	c.Scan()

	// varName
	c.TagMust(tagIdentifier)
	c.PrintLine()
	name := c.value
	c.Scan()

	c.symbolTable.Define(name, typeName, "local")

	// (, varName)*
	for c.ValueIs(",") {
		// ,
		c.PrintLine()
		c.Scan()
		// varName
		c.TagMust(tagIdentifier)
		c.PrintLine()
		name = c.value
		c.symbolTable.Define(name, typeName, "local")
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
	name := c.value
	c.PrintLine()

	kind := c.symbolTable.KindOf(name)
	typeName := c.symbolTable.TypeOf(name)
	number := c.symbolTable.IndexOf(name)
	hasIndex := false

	// ( [ expression ] )?
	c.Scan()
	if c.ValueIs("[") {
		hasIndex = true

		// [
		c.PrintLine()
		c.Scan()

		// expression
		c.CompileExpression()
		c.vmWriter.WritePop("temp", 1)

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

	if typeName == "Array" && hasIndex {
		// calc addr + index
		c.vmWriter.WritePush(kindToSegment[kind], number)
		c.vmWriter.WritePush("temp", 1)
		c.vmWriter.WriteArithmetic("add")
		// set pointer
		c.vmWriter.WritePop("pointer", 1)
		// insert that
		c.vmWriter.WritePop("that", 0)
	} else {
		// insert var
		c.vmWriter.WritePop(kindToSegment[kind], number)
	}

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

	c.vmWriter.WriteReturn()
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

	c.labelCount += 1
	l1 := fmt.Sprintf("%s.label.%d", c.className, c.labelCount)
	c.labelCount += 1
	l2 := fmt.Sprintf("%s.label.%d", c.className, c.labelCount)

	c.vmWriter.WriteArithmetic("not")
	c.vmWriter.WriteIf(l1)

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

	c.vmWriter.WriteGoto(l2)
	c.vmWriter.WriteLabel(l1)

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

	c.vmWriter.WriteLabel(l2)

	// close ifStatement
	c.indent -= 1
	c.Println("</ifStatement>")
}

func (c *Compiler) CompileWhileStatement() {
	// whileStatement
	c.Println("<whileStatement>")
	c.indent += 1

	c.labelCount += 1
	l1 := fmt.Sprintf("%s.label.%d", c.className, c.labelCount)
	c.labelCount += 1
	l2 := fmt.Sprintf("%s.label.%d", c.className, c.labelCount)

	// while
	c.ValueMust("while")
	c.PrintLine()
	c.Scan()

	// (
	c.ValueMust("(")
	c.PrintLine()
	c.Scan()

	c.vmWriter.WriteLabel(l1)

	// expression
	c.CompileExpression()

	c.vmWriter.WriteArithmetic("not")
	c.vmWriter.WriteIf(l2)

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

	c.vmWriter.WriteGoto(l1)
	c.vmWriter.WriteLabel(l2)

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
		op := c.value
		c.Scan()
		// term
		c.CompileTerm()

		switch op {
		case "+":
			c.vmWriter.WriteArithmetic("add")
		case "-":
			c.vmWriter.WriteArithmetic("sub")
		case "*":
			c.vmWriter.WriteCall("Math.multiply", 2)
		case "/":
			c.vmWriter.WriteCall("Math.divide", 2)
		case "&":
			c.vmWriter.WriteArithmetic("and")
		case "|":
			c.vmWriter.WriteArithmetic("or")
		case "<":
			c.vmWriter.WriteArithmetic("lt")
		case ">":
			c.vmWriter.WriteArithmetic("gt")
		case "=":
			c.vmWriter.WriteArithmetic("eq")
		}
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
		v := c.value
		c.Scan()

		i, err := strconv.Atoi(v)
		if err != nil {
			log.Fatal(err)
		}
		c.vmWriter.WritePush("constant", i)
	} else if c.TagIs("stringConstant") {
		// stringConstant
		c.PrintLine()
		s := c.value
		c.Scan()

		c.vmWriter.WritePush("constant", len(s)) // String.new arg 0
		c.vmWriter.WriteCall("String.new", 1) // new String. stack head is string object.
		for _, char := range s {
			c.vmWriter.WritePush("constant", int(char)) // String.appendChar arg 1
			c.vmWriter.WriteCall("String.appendChar", 2)
			// c.vmWriter.WritePop("temp", 0) // pop void return
		}

	} else if c.TagIs("keyword") && c.ValueIs("true", "false", "null", "this") {
		// true | false | null | this
		c.PrintLine()
		v := c.value
		c.Scan()

		switch v {
		case "true":
			c.vmWriter.WritePush("constant", 1)
			c.vmWriter.WriteArithmetic("neg")
		case "false", "null":
			c.vmWriter.WritePush("constant", 0)
		case "this":
			c.vmWriter.WritePush("pointer", 0)
		}
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
		op := c.value
		c.Scan()
		// term
		c.CompileTerm()

		switch op {
		case "-":
			c.vmWriter.WriteArithmetic("neg")
		case "~":
			c.vmWriter.WriteArithmetic("not")
		}
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
	name := c.value
	typeName := c.symbolTable.TypeOf(name)
	kind := c.symbolTable.KindOf(name)
	number := c.symbolTable.IndexOf(name)
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

		c.vmWriter.WritePush(kindToSegment[kind], number)
		c.vmWriter.WriteArithmetic("add")
		c.vmWriter.WritePop("pointer", 1)
		c.vmWriter.WritePush("that", 0)

	} else if c.ValueIs(".") {
		// (className | varName) . subroutineName ( expressionList )

		// .
		c.PrintLine()
		c.Scan()

		// subroutineName
		c.TagMust(tagIdentifier)
		c.PrintLine()
		subroutineName := c.value
		c.Scan()

		// set "this" object to stack
		isMethod := false
		className := ""
		if kind == "none" {
			className = name
		} else {
			isMethod = true
			className = typeName
			c.vmWriter.WritePush(kindToSegment[kind], number)
		}

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

		argCount := c.expressionListCount
		if isMethod {
			argCount += 1
		}
		c.vmWriter.WriteCall(fmt.Sprintf("%s.%s", className, subroutineName), argCount)

	} else if c.ValueIs("(") {
		// subroutineName ( expressionList )

		isMethod := false
		for _, e := range c.symbolTable.subroutine {
			// called method without "this"
			if e.name == name && e.kind == "method" {
				isMethod = true
				c.vmWriter.WritePush("pointer", 0)
				break
			}
		}

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

		argCount := c.expressionListCount
		if isMethod {
			argCount += 1
		}
		c.vmWriter.WriteCall(fmt.Sprintf("%s.%s", c.className, name), argCount)

	} else {
		// varName
		// no op

		c.vmWriter.WritePush(kindToSegment[kind], number)
	}
}

// ( expression (, expression)* )?
func (c *Compiler) CompileExpressionList() {
	// expressionList
	c.Println("<expressionList>")
	c.indent += 1

	count := 0

	if c.IsTerm() {

		// expression
		c.CompileExpression()

		count += 1

		for c.ValueIs(",") {
			// ,
			c.PrintLine()
			c.Scan()

			// expression
			c.CompileExpression()

			count += 1
		}

	}

	// close expressionList
	c.indent -= 1
	c.Println("</expressionList>")

	c.expressionListCount = count
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
