package main

import "testing"

func TestCode_Dest(t *testing.T) {
	c := NewCode()

	samples := []struct {
		In  string
		Out uint16
	}{
		{"", 0},
		{"M", destMBit},
		{"D", destDBit},
		{"MD", destMBit | destDBit},
		{"A", destABit},
		{"AM", destABit | destMBit},
		{"AD", destABit | destDBit},
		{"AMD", destABit | destMBit | destDBit},
	}
	for _, s := range samples {
		out := c.Dest(s.In)
		if out != s.Out {
			t.Errorf(`Sample: %#v, Out: %d`, s, out)
		}
		t.Logf(`Sample: %#v, Out: %d`, s, out)
	}
}
