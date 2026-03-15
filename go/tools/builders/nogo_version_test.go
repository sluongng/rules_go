package main

import "testing"

func TestTrimGoPatchVersion(t *testing.T) {
	testCases := []struct {
		name      string
		goVersion string
		want      string
	}{
		{name: "patch with prefix", goVersion: "go1.20.14", want: "go1.20"},
		{name: "patch without prefix", goVersion: "1.20.14", want: "1.20"},
		{name: "major minor only", goVersion: "go1.21", want: "go1.21"},
		{name: "suffix without patch", goVersion: "go1.21rc1", want: "go1.21rc1"},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := trimGoPatchVersion(tc.goVersion); got != tc.want {
				t.Fatalf("trimGoPatchVersion(%q) = %q, want %q", tc.goVersion, got, tc.want)
			}
		})
	}
}
