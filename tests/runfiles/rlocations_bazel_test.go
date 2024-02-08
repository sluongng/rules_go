// Copyright 2019 The Bazel Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rlocations_test

import (
	"regexp"
	"runtime"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel_testing"
)

func TestMain(m *testing.M) {
	bazel_testing.TestMain(m, bazel_testing.Args{
		Main: `
-- files/a --
-- files/b --
-- files/BUILD.bazel --
filegroup(
    name = "files",
    srcs = ["a", "b"],
    visibility = ["//visibility:public"],
)
-- BUILD.bazel --
load("@io_bazel_rules_go//go:def.bzl", "go_binary")

go_binary(
    name = "main",
    data = ["//files"],
    srcs = ["main.go"],
    x_defs = {
        "filesRunfilePath": "$(rlocationpaths //files)",
    },
    deps = [
        "@io_bazel_rules_go//go/runfiles",
    ],
)
-- main.go --
package main

import (
	"fmt"
	"sort"

	"github.com/bazelbuild/rules_go/go/runfiles"
)

// set by x_defs in BUILD.bazel
var filesRunfilePath string

func main() {
	paths, err := runfiles.Rlocations(filesRunfilePath)
	if err != nil {
		panic(err)
	}

	keys := make([]string, 0, len(paths))
	for k := range paths {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		fmt.Printf("file: %q - %q\n", k, paths[k])
	}
}
`,
	})
}

var unixExpectedOutputs = []*regexp.Regexp{
	regexp.MustCompile(`file: "(_main|__main__)/files/a" - "/.*/main/files/a"\n`),
	regexp.MustCompile(`file: "(_main|__main__)/files/b" - "/.*/main/files/b"\n`),
}

var windowsExpectedOutputs = []*regexp.Regexp{
	regexp.MustCompile(`file: "(_main|__main__)/files/a" - ".:\\\\.*\\\\main\\\\files\\\\a"\n`),
	regexp.MustCompile(`file: "(_main|__main__)/files/b" - ".:\\\\.*\\\\main\\\\files\\\\b"\n`),
}

func TestCurrentRepository(t *testing.T) {
	out, err := bazel_testing.BazelOutput("run", "//:main")
	if err != nil {
		t.Fatal(err)
	}

	expectedOutputs := unixExpectedOutputs
	if runtime.GOOS == "windows" {
		expectedOutputs = windowsExpectedOutputs
	}

	for _, expected := range expectedOutputs {
		if !expected.Match(out) {
			t.Fatalf("got: %q,\n---\n want: %q", string(out), expected.String())
		}
	}
}
