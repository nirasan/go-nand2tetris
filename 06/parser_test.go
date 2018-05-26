package main

import "testing"

func TestNewParser(t *testing.T) {
	p, err := NewParser(`test/MaxL.asm`)
	if err != nil {
		t.Fatal(err)
	}
	for p.HasMoreCommands() {
		t.Log(`=====`)
		t.Log(`CurrentLine: `, p.CurrentLine)
		t.Log(`CommandType: `, p.CommandType())
		t.Log(`Symbol: `, p.Symbol())
		t.Log(`Dest: `, p.Dest())
		t.Log(`Comp: `, p.Comp())
		t.Log(`Jump: `, p.Jump())
	}
}
