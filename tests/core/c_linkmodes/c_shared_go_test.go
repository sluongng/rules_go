package c_linkmodes

/*
#include "tests/core/c_linkmodes/adder_shared.h"
*/
import "C"

import "testing"

func TestSharedGoAdd(t *testing.T) {
	got := int(C.GoAdd(40, 2))
	if got != 42 {
		t.Fatalf("GoAdd(40, 2) = %d, want 42", got)
	}
}
