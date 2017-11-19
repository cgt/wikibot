package main

import (
	"sort"
	"testing"
)

func TestDifference(t *testing.T) {
	a := []string{"a", "b", "c", "d"}
	b := []string{"a", "c"}
	sort.Strings(a)
	sort.Strings(b)

	c := difference(a, b)
	if len(c) != 2 || len(c) == 2 && (c[0] != "b" || c[1] != "d") {
		t.Errorf("got=%v expected=%v", c, []string{"b", "d"})
	}
}
