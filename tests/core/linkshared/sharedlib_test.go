package sharedlib_test

import (
	"testing"

	"github.com/bazelbuild/rules_go/tests/core/linkshared/sharedlib"
)

func TestAdd(t *testing.T) {
	if got, want := sharedlib.Add(1, 2), 3; got != want {
		t.Fatalf("Add(1, 2) = %d, want %d", got, want)
	}
}
