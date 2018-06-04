package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"fmt"
)

func main() {
	filename := os.Args[1]
	stat, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}

	var files []*os.File
	if stat.IsDir() {
		list, err := ioutil.ReadDir(filename)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range list {
			if !strings.HasSuffix(f.Name(), ".jack") {
				continue
			}
			ff, err := os.Open(filename + f.Name())
			if err != nil {
				log.Fatal(err)
			}
			files = append(files, ff)
		}
	} else {
		f, err := os.Open(filename)
		if err != nil {
			log.Fatal(err)
		}
		files = append(files, f)
	}

	for _, f := range files {
		log.Printf("FILE: %s", f.Name())
		tokenizer := NewTokenizer(f)

		w, err := os.Create(f.Name() + ".xml")
		if err != nil {
			panic(err)
		}

		fmt.Fprint(w, "<tokens>\n")

		for tokenizer.HasMoreTokens() {
			log.Printf("%#v", tokenizer)
			switch tokenizer.TokenType() {
			case TOKEN_SYMBOL:
				fmt.Fprintf(w, "<symbol> %s </symbol>\n", tokenizer.Symbol())
			case TOKEN_KEYWORD:
				fmt.Fprintf(w, "<keyword> %s </keyword>\n", tokenizer.KeyWord())
			case TOKEN_INT_CONST:
				fmt.Fprintf(w, "<integerConstant> %d </integerConstant>\n", tokenizer.IntVal())
			case TOKEN_STRING_CONST:
				fmt.Fprintf(w, "<stringConstant> %s </stringConstant>\n", tokenizer.StringVal())
			case TOKEN_IDENTIFIER:
				fmt.Fprintf(w, "<identifier> %s </identifier>\n", tokenizer.Identifier())
			}
		}

		fmt.Fprint(w, "</tokens>\n")

		w.Close()
		f.Close()
	}
}
