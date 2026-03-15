//go:build go1.18 && !go1.21
// +build go1.18,!go1.21

package main

import "testing"

func TestNormalizeGoVersionPre121(t *testing.T) {
	if got := normalizeGoVersion("1.20.14"); got != "go1.20" {
		t.Fatalf("normalizeGoVersion(1.20.14) = %q, want %q", got, "go1.20")
	}
}
