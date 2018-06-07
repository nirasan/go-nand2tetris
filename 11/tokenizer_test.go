package main

import (
	"bytes"
	"os"
	"testing"
)

func TestTokenizer_HasMoreTokens(t *testing.T) {
	tk := NewTokenizer(bytes.NewBufferString(`
	// comment1
	/* Foo class */
	class Foo {
	  function void main() {
	    var int i;
	    var String s;
	    let i = 100;
	    let s = "hello world";
	  }
	}
	`))
	for tk.HasMoreTokens() {
		t.Logf("Token: %#v, TokenType: %s", tk.token, tk.tokenType)
	}
}

func TestTokenizer_HasMoreTokens2(t *testing.T) {
	f, err := os.Open("test/ArrayTest/Main.jack")
	if err != nil {
		t.Fatal(err)
	}
	tk := NewTokenizer(f)
	for tk.HasMoreTokens() {
		t.Logf("Token: %#v, TokenType: %s", tk.token, tk.tokenType)
	}
}
