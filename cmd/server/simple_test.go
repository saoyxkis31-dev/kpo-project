package main

import "testing"

func TestSimple(t *testing.T) {
	x := 10

	if x != 10 {
		t.Errorf("x ne 10")
	}
}
