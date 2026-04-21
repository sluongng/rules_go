// Copyright 2026 The Bazel Authors. All rights reserved.
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

package buildinfo_test

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/bazelbuild/rules_go/go/tools/bazel_testing"
)

const (
	defaultVCSRevision = "abc123"
	defaultVCSTime     = "2026-04-21T17:47:02+07:00"
	defaultVCSStatus   = "Modified"
)

func TestMain(m *testing.M) {
	bazel_testing.TestMain(m, bazel_testing.Args{
		Main: `
-- BUILD.bazel --
load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")
load(":custom_binary.bzl", "custom_binary", "custom_multi_binary")

package_metadata(
    name = "main_package_metadata",
    purl = "pkg:golang/example.com/main@v1.0.0",
)

go_binary(
    name = "with_dep",
    srcs = ["with_dep.go"],
    applicable_licenses = [":main_package_metadata"],
    deps = [
        "@com_example_versionless//:go_default_library",  # Declared but not imported.
        "@com_github_google_go_cmp//cmp:go_default_library",
    ],
)

go_library(
    name = "embedded_main",
    srcs = ["with_dep.go"],
    importpath = "example.com/main/cmd/embedded_wrapper",
    applicable_licenses = [":main_package_metadata"],
    deps = ["@com_github_google_go_cmp//cmp:go_default_library"],
)

go_binary(
    name = "embedded_wrapper",
    embed = [":embedded_main"],
)

go_test(
    name = "with_dep_test",
    srcs = ["with_dep_test.go"],
    applicable_licenses = [":main_package_metadata"],
    deps = ["@com_github_google_go_cmp//cmp:go_default_library"],
)

go_binary(
    name = "with_stable_xdef",
    srcs = [
        "stdlib_only.go",
        "build_user.go",
    ],
    applicable_licenses = [":main_package_metadata"],
    x_defs = {
        "BuildUser": "{BUILD_USER}",
    },
)

go_binary(
    name = "stdlib_only",
    srcs = ["stdlib_only.go"],
)

custom_binary(
    name = "custom_stamped",
    srcs = ["stdlib_only.go"],
    importpath = "example.com/main/cmd/custom_stamped",
    applicable_licenses = [":main_package_metadata"],
)

custom_multi_binary(
    name = "custom_multi_output",
    srcs = ["stdlib_only.go"],
    importpath = "example.com/main/cmd/custom_multi_output",
    applicable_licenses = [":main_package_metadata"],
)

go_binary(
    name = "with_versionless_dep",
    srcs = ["with_versionless_dep.go"],
    deps = ["@com_example_versionless//:go_default_library"],
    applicable_licenses = [":main_package_metadata"],
)

go_binary(
    name = "with_vendored_dep",
    srcs = ["with_vendored_dep.go"],
    deps = ["//third_party/vendored:vendored"],
)

go_binary(
    name = "with_external_src",
    srcs = ["@com_example_external_sources//:external_main.go"],
    applicable_licenses = [":main_package_metadata"],
)

go_binary(
    name = "embedded_external_wrapper",
    embed = ["@com_example_external_embedded//:external_embedded_main"],
)

go_binary(
    name = "metadata_external_wrapper",
    srcs = ["stdlib_only.go"],
    applicable_licenses = ["@com_example_external_metadata//:external_metadata"],
)

go_library(
    name = "localdep",
    srcs = ["localdep.go"],
    importpath = "example.com/mainmodule/localdep",
    applicable_licenses = [":main_package_metadata"],
)

go_library(
    name = "subject",
    srcs = ["subject.go"],
    importpath = "example.com/mainmodule/subject",
    applicable_licenses = [":main_package_metadata"],
)

go_test(
    name = "subject_test",
    srcs = ["subject_test.go"],
    embed = [":subject"],
)

go_test(
    name = "subject_stamped_test",
    srcs = ["subject_stamped_test.go"],
    embed = [":subject"],
)

go_test(
    name = "subject_external_metadata_test",
    srcs = ["subject_external_metadata_test.go"],
    embed = [":subject"],
    applicable_licenses = ["@com_example_external_metadata//:external_metadata"],
)

-- custom_binary.bzl --
load("@io_bazel_rules_go//go:def.bzl", "go_context", "go_rule", "new_go_info")

def _custom_binary_impl(ctx):
    go = go_context(ctx, maybe_needs_cc_toolchain = False)
    source = new_go_info(
        go,
        ctx.attr,
        importable = False,
        is_main = True,
    )
    archive, executable, runfiles = go.binary(
        go,
        ctx.label.name,
        source,
        [],
        [],
        ctx.version_file,
        ctx.info_file,
    )
    return [
        archive,
        DefaultInfo(
            files = depset([executable]),
            runfiles = runfiles,
            executable = executable,
        ),
    ]

custom_binary = go_rule(
    _custom_binary_impl,
    executable = True,
    attrs = {
        "srcs": attr.label_list(allow_files = [".go"]),
        "importpath": attr.string(mandatory = True),
        "_go_context_data": attr.label(default = "@io_bazel_rules_go//:go_context_data"),
    },
)

def _custom_multi_binary_impl(ctx):
    go = go_context(ctx, maybe_needs_cc_toolchain = False)
    source = new_go_info(
        go,
        ctx.attr,
        importable = False,
        is_main = True,
    )
    archive = go.archive(go, source)
    executables = [
        go.actions.declare_file("a/tool"),
        go.actions.declare_file("b/tool"),
    ]
    for executable in executables:
        go.link(
            go,
            archive,
            [],
            executable,
            [],
            ctx.version_file,
            ctx.info_file,
        )
    return [archive, DefaultInfo(files = depset(executables))]

custom_multi_binary = go_rule(
    _custom_multi_binary_impl,
    attrs = {
        "srcs": attr.label_list(allow_files = [".go"]),
        "importpath": attr.string(mandatory = True),
        "_go_context_data": attr.label(default = "@io_bazel_rules_go//:go_context_data"),
    },
)

-- with_dep.go --
package main

import (
    "encoding/json"
    "os"
    "runtime/debug"

    "github.com/google/go-cmp/cmp"
)

type dep struct {
    Path    string ` + "`json:\"path\"`" + `
    Version string ` + "`json:\"version\"`" + `
}

type output struct {
    OK          bool              ` + "`json:\"ok\"`" + `
    Path        string            ` + "`json:\"path\"`" + `
    MainPath    string            ` + "`json:\"main_path\"`" + `
    MainVersion string            ` + "`json:\"main_version\"`" + `
    Settings    map[string]string ` + "`json:\"settings\"`" + `
    Deps        []dep             ` + "`json:\"deps\"`" + `
}

func buildSettings(info *debug.BuildInfo) map[string]string {
    settings := make(map[string]string, len(info.Settings))
    for _, setting := range info.Settings {
        settings[setting.Key] = setting.Value
    }
    return settings
}

func main() {
    _ = cmp.Equal("same", "same")

    info, ok := debug.ReadBuildInfo()
    out := output{OK: ok}
    if info != nil {
        out.Path = info.Path
        out.MainPath = info.Main.Path
        out.MainVersion = info.Main.Version
        out.Settings = buildSettings(info)
        for _, module := range info.Deps {
            out.Deps = append(out.Deps, dep{Path: module.Path, Version: module.Version})
        }
    }
    _ = json.NewEncoder(os.Stdout).Encode(out)
}

-- with_dep_test.go --
package buildinfo

import (
    "runtime/debug"
    "testing"

    "github.com/google/go-cmp/cmp"
)

func TestBuildInfoDeps(t *testing.T) {
    _ = cmp.Equal("same", "same")

    info, ok := debug.ReadBuildInfo()
    if !ok {
        t.Fatal("ReadBuildInfo returned ok=false")
    }
    if info.Main.Path != "example.com/main" || info.Main.Version != "v1.0.0" {
        t.Fatalf("got Main %+v; want example.com/main@v1.0.0", info.Main)
    }

    foundCmp := false
    for _, dep := range info.Deps {
        if dep.Path == "example.com/main" {
            t.Fatalf("test target metadata was included as a dependency: %+v", info.Deps)
        }
        if dep.Path == "github.com/google/go-cmp" && dep.Version == "v0.6.0" {
            foundCmp = true
        }
    }
    if !foundCmp {
        t.Fatalf("missing github.com/google/go-cmp@v0.6.0 in %+v", info.Deps)
    }
}

-- localdep.go --
package localdep

func Name() string {
    return "localdep"
}

-- embedded_main.go --
package main

import (
    "encoding/json"
    "os"
    "runtime/debug"
)

func main() {
    info, ok := debug.ReadBuildInfo()
    _ = json.NewEncoder(os.Stdout).Encode(struct {
        OK          bool   ` + "`json:\"ok\"`" + `
        Path        string ` + "`json:\"path\"`" + `
        MainPath    string ` + "`json:\"main_path\"`" + `
        MainVersion string ` + "`json:\"main_version\"`" + `
    }{
        OK: ok,
        Path: info.Path,
        MainPath: info.Main.Path,
        MainVersion: info.Main.Version,
    })
}

-- subject.go --
package subject

func Name() string {
    return "subject"
}

-- subject_test.go --
package subject

import (
    "runtime/debug"
    "testing"

    "github.com/google/go-cmp/cmp"
)

func TestBuildInfoDeps(t *testing.T) {
    _ = cmp.Equal("same", "same")

    info, ok := debug.ReadBuildInfo()
    if !ok {
        t.Fatal("ReadBuildInfo returned ok=false")
    }
	if info.Main.Path != "example.com/main" || info.Main.Version != "v1.0.0" {
		t.Fatalf("got Main %+v; want example.com/main@v1.0.0", info.Main)
	}

    foundCmp := false
    for _, dep := range info.Deps {
        if dep.Path == "example.com/main" {
            t.Fatalf("test target metadata was included as a dependency: %+v", info.Deps)
        }
        if dep.Path == "github.com/google/go-cmp" && dep.Version == "v0.6.0" {
            foundCmp = true
        }
    }
    if !foundCmp {
        t.Fatalf("missing github.com/google/go-cmp@v0.6.0 in %+v", info.Deps)
    }
}

-- subject_stamped_test.go --
package subject

import (
    "runtime/debug"
    "testing"
)

func buildSettings(info *debug.BuildInfo) map[string]string {
    out := make(map[string]string, len(info.Settings))
    for _, setting := range info.Settings {
        out[setting.Key] = setting.Value
    }
    return out
}

func TestBuildInfoVCSSettings(t *testing.T) {
    info, ok := debug.ReadBuildInfo()
    if !ok {
        t.Fatalf("ReadBuildInfo returned ok=false: %+v", info)
    }

    settings := buildSettings(info)
    if settings["vcs"] != "git" {
        t.Fatalf("got vcs %q; want %q", settings["vcs"], "git")
    }
    if settings["vcs.revision"] != "abc123" {
        t.Fatalf("got vcs.revision %q; want %q", settings["vcs.revision"], "abc123")
    }
    if settings["vcs.time"] != "2026-04-21T10:47:02Z" {
        t.Fatalf("got vcs.time %q; want %q", settings["vcs.time"], "2026-04-21T10:47:02Z")
    }
    if settings["vcs.modified"] != "true" {
        t.Fatalf("got vcs.modified %q; want %q", settings["vcs.modified"], "true")
    }
}

-- subject_external_metadata_test.go --
package subject

import (
    "runtime/debug"
    "testing"
)

func buildSettings(info *debug.BuildInfo) map[string]string {
    out := make(map[string]string, len(info.Settings))
    for _, setting := range info.Settings {
        out[setting.Key] = setting.Value
    }
    return out
}

func TestBuildInfoExternalMetadataOmitsVCSSettings(t *testing.T) {
    info, ok := debug.ReadBuildInfo()
    if !ok {
        t.Fatalf("ReadBuildInfo returned ok=false: %+v", info)
    }
    if info.Main.Path != "example.com/externalmetadata" || info.Main.Version != "v7.8.9" {
        t.Fatalf("got Main %+v; want example.com/externalmetadata@v7.8.9", info.Main)
    }

    settings := buildSettings(info)
    for _, key := range []string{"vcs", "vcs.revision", "vcs.time", "vcs.modified"} {
        if value := settings[key]; value != "" {
            t.Fatalf("got %s=%q in external metadata stamped build info; want empty", key, value)
        }
    }
}

-- stdlib_only.go --
package main

import (
    "encoding/json"
    "os"
    "runtime/debug"
)

type output struct {
	OK          bool              ` + "`json:\"ok\"`" + `
	Path        string            ` + "`json:\"path\"`" + `
	MainPath    string            ` + "`json:\"main_path\"`" + `
	MainVersion string            ` + "`json:\"main_version\"`" + `
	DepCount    int               ` + "`json:\"dep_count\"`" + `
	Settings    map[string]string ` + "`json:\"settings\"`" + `
}

func buildSettings(info *debug.BuildInfo) map[string]string {
	settings := make(map[string]string, len(info.Settings))
	for _, setting := range info.Settings {
		settings[setting.Key] = setting.Value
	}
	return settings
}

func main() {
    info, ok := debug.ReadBuildInfo()
    out := output{OK: ok}
    if info != nil {
        out.Path = info.Path
        out.MainPath = info.Main.Path
		out.MainVersion = info.Main.Version
		out.DepCount = len(info.Deps)
		out.Settings = buildSettings(info)
    }
    _ = json.NewEncoder(os.Stdout).Encode(out)
}

-- build_user.go --
package main

var BuildUser = "unset"

-- with_versionless_dep.go --
package main

import (
    "encoding/json"
    "os"
    "runtime/debug"

    versionless "example.com/versionless"
)

type dep struct {
    Path    string ` + "`json:\"path\"`" + `
    Version string ` + "`json:\"version\"`" + `
}

type output struct {
    OK          bool   ` + "`json:\"ok\"`" + `
    Path        string ` + "`json:\"path\"`" + `
    MainPath    string ` + "`json:\"main_path\"`" + `
    MainVersion string ` + "`json:\"main_version\"`" + `
    Deps        []dep  ` + "`json:\"deps\"`" + `
}

func main() {
    _ = versionless.Name()

    info, ok := debug.ReadBuildInfo()
    out := output{OK: ok}
    if info != nil {
        out.Path = info.Path
        out.MainPath = info.Main.Path
        out.MainVersion = info.Main.Version
        for _, module := range info.Deps {
            out.Deps = append(out.Deps, dep{Path: module.Path, Version: module.Version})
        }
    }
    _ = json.NewEncoder(os.Stdout).Encode(out)
}

-- with_vendored_dep.go --
package main

import (
    "encoding/json"
    "os"
    "runtime/debug"

    "example.com/vendored"
)

type dep struct {
    Path    string ` + "`json:\"path\"`" + `
    Version string ` + "`json:\"version\"`" + `
}

type output struct {
    OK   bool  ` + "`json:\"ok\"`" + `
    Deps []dep ` + "`json:\"deps\"`" + `
}

func main() {
    _ = vendored.Name()

    info, ok := debug.ReadBuildInfo()
    out := output{OK: ok}
    if info != nil {
        for _, module := range info.Deps {
            out.Deps = append(out.Deps, dep{Path: module.Path, Version: module.Version})
        }
    }
    _ = json.NewEncoder(os.Stdout).Encode(out)
}

-- third_party/vendored/BUILD.bazel --
load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "package_metadata",
    purl = "pkg:golang/example.com/vendored@v1.2.3",
)

go_library(
    name = "vendored",
    srcs = ["vendored.go"],
    importpath = "example.com/vendored",
    applicable_licenses = [":package_metadata"],
    visibility = ["//visibility:public"],
)

-- third_party/vendored/vendored.go --
package vendored

func Name() string {
    return "vendored"
}

-- external/tool/BUILD.bazel --
load("@io_bazel_rules_go//go:def.bzl", "go_binary")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "package_metadata",
    purl = "pkg:golang/example.com/main@v1.0.0",
)

go_binary(
    name = "bin",
    srcs = ["main.go"],
    applicable_licenses = [":package_metadata"],
)

-- external/tool/main.go --
package main

func main() {}

-- deps/com_github_google_go_cmp/MODULE.bazel --
module(name = "com_github_google_go_cmp")

bazel_dep(name = "rules_go", repo_name = "io_bazel_rules_go")
bazel_dep(name = "package_metadata", version = "0.0.5")

-- deps/com_github_google_go_cmp/BUILD.bazel --
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "package_metadata",
    purl = "pkg:golang/github.com/google/go-cmp@v0.6.0",
    visibility = ["//:__subpackages__"],
)

-- deps/com_github_google_go_cmp/cmp/BUILD.bazel --
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "cmp",
    srcs = ["cmp.go"],
    importpath = "github.com/google/go-cmp/cmp",
    applicable_licenses = ["//:package_metadata"],
    visibility = ["//visibility:public"],
)

alias(
    name = "go_default_library",
    actual = ":cmp",
    visibility = ["//visibility:public"],
)

-- deps/com_github_google_go_cmp/cmp/cmp.go --
package cmp

func Equal(x, y string) bool {
    return x == y
}

-- deps/com_example_versionless/MODULE.bazel --
module(name = "com_example_versionless")

bazel_dep(name = "rules_go", repo_name = "io_bazel_rules_go")
bazel_dep(name = "package_metadata", version = "0.0.5")

-- deps/com_example_versionless/BUILD.bazel --
load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "package_metadata",
    purl = "pkg:golang/example.com/versionless",
    visibility = ["//:__subpackages__"],
)

go_library(
    name = "versionless",
    srcs = ["versionless.go"],
    importpath = "example.com/versionless",
    applicable_licenses = [":package_metadata"],
    visibility = ["//visibility:public"],
)

alias(
    name = "go_default_library",
    actual = ":versionless",
    visibility = ["//visibility:public"],
)

-- deps/com_example_versionless/versionless.go --
package versionless

func Name() string {
    return "versionless"
}

-- deps/com_example_external_sources/MODULE.bazel --
module(name = "com_example_external_sources")

-- deps/com_example_external_sources/BUILD.bazel --
exports_files(["external_main.go"])

-- deps/com_example_external_sources/external_main.go --
package main

import (
    "encoding/json"
    "os"
    "runtime/debug"
)

type output struct {
    OK          bool              ` + "`json:\"ok\"`" + `
    Path        string            ` + "`json:\"path\"`" + `
    MainPath    string            ` + "`json:\"main_path\"`" + `
    MainVersion string            ` + "`json:\"main_version\"`" + `
    Settings    map[string]string ` + "`json:\"settings\"`" + `
}

func buildSettings(info *debug.BuildInfo) map[string]string {
    out := make(map[string]string, len(info.Settings))
    for _, setting := range info.Settings {
        out[setting.Key] = setting.Value
    }
    return out
}

func main() {
    info, ok := debug.ReadBuildInfo()
    out := output{OK: ok}
    if info != nil {
        out.Path = info.Path
        out.MainPath = info.Main.Path
        out.MainVersion = info.Main.Version
        out.Settings = buildSettings(info)
    }
    _ = json.NewEncoder(os.Stdout).Encode(out)
}

-- deps/com_example_external_binary/MODULE.bazel --
module(name = "com_example_external_binary")

bazel_dep(name = "rules_go", repo_name = "io_bazel_rules_go")
bazel_dep(name = "package_metadata", version = "0.0.5")

-- deps/com_example_external_binary/BUILD.bazel --
load("@io_bazel_rules_go//go:def.bzl", "go_binary")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "external_binary_package_metadata",
    purl = "pkg:golang/example.com/externaltool@v9.9.9",
    visibility = ["//:__subpackages__"],
)

go_binary(
    name = "tool",
    srcs = ["main.go"],
    applicable_licenses = [":external_binary_package_metadata"],
)

-- deps/com_example_external_binary/main.go --
package main

import (
    "encoding/json"
    "os"
    "runtime/debug"
)

type output struct {
    OK          bool              ` + "`json:\"ok\"`" + `
    Path        string            ` + "`json:\"path\"`" + `
    MainPath    string            ` + "`json:\"main_path\"`" + `
    MainVersion string            ` + "`json:\"main_version\"`" + `
    Settings    map[string]string ` + "`json:\"settings\"`" + `
}

func buildSettings(info *debug.BuildInfo) map[string]string {
    out := make(map[string]string, len(info.Settings))
    for _, setting := range info.Settings {
        out[setting.Key] = setting.Value
    }
    return out
}

func main() {
    info, ok := debug.ReadBuildInfo()
    out := output{OK: ok}
    if info != nil {
        out.Path = info.Path
        out.MainPath = info.Main.Path
        out.MainVersion = info.Main.Version
        out.Settings = buildSettings(info)
    }
    _ = json.NewEncoder(os.Stdout).Encode(out)
}

-- deps/com_example_external_embedded/MODULE.bazel --
module(name = "com_example_external_embedded")

bazel_dep(name = "rules_go", repo_name = "io_bazel_rules_go")
bazel_dep(name = "package_metadata", version = "0.0.5")

-- deps/com_example_external_embedded/BUILD.bazel --
load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "external_embedded_package_metadata",
    purl = "pkg:golang/example.com/externalembedded@v4.5.6",
    visibility = ["//:__subpackages__"],
)

go_library(
    name = "external_embedded_main",
    srcs = ["main.go"],
    importpath = "example.com/externalembedded/cmd/tool",
    applicable_licenses = [":external_embedded_package_metadata"],
    visibility = ["//visibility:public"],
)

-- deps/com_example_external_embedded/main.go --
package main

import (
    "encoding/json"
    "os"
    "runtime/debug"
)

type output struct {
    OK          bool              ` + "`json:\"ok\"`" + `
    Path        string            ` + "`json:\"path\"`" + `
    MainPath    string            ` + "`json:\"main_path\"`" + `
    MainVersion string            ` + "`json:\"main_version\"`" + `
    Settings    map[string]string ` + "`json:\"settings\"`" + `
}

func buildSettings(info *debug.BuildInfo) map[string]string {
    out := make(map[string]string, len(info.Settings))
    for _, setting := range info.Settings {
        out[setting.Key] = setting.Value
    }
    return out
}

func main() {
    info, ok := debug.ReadBuildInfo()
    out := output{OK: ok}
    if info != nil {
        out.Path = info.Path
        out.MainPath = info.Main.Path
        out.MainVersion = info.Main.Version
        out.Settings = buildSettings(info)
    }
    _ = json.NewEncoder(os.Stdout).Encode(out)
}

-- deps/com_example_external_metadata/MODULE.bazel --
module(name = "com_example_external_metadata")

bazel_dep(name = "package_metadata", version = "0.0.5")

-- deps/com_example_external_metadata/BUILD.bazel --
load("@package_metadata//rules:package_metadata.bzl", "package_metadata")

package_metadata(
    name = "external_metadata",
    purl = "pkg:golang/example.com/externalmetadata@v7.8.9",
    visibility = ["//visibility:public"],
)
`,
		ModuleFileSuffix: `
bazel_dep(name = "package_metadata", version = "0.0.5")
bazel_dep(name = "com_github_google_go_cmp")
bazel_dep(name = "com_example_versionless")
bazel_dep(name = "com_example_external_sources")
bazel_dep(name = "com_example_external_binary")
bazel_dep(name = "com_example_external_embedded")
bazel_dep(name = "com_example_external_metadata")

local_path_override(
    module_name = "com_github_google_go_cmp",
    path = "deps/com_github_google_go_cmp",
)

local_path_override(
    module_name = "com_example_versionless",
    path = "deps/com_example_versionless",
)

local_path_override(
    module_name = "com_example_external_sources",
    path = "deps/com_example_external_sources",
)

local_path_override(
    module_name = "com_example_external_binary",
    path = "deps/com_example_external_binary",
)

local_path_override(
    module_name = "com_example_external_embedded",
    path = "deps/com_example_external_embedded",
)

local_path_override(
    module_name = "com_example_external_metadata",
    path = "deps/com_example_external_metadata",
)
`,
		SetUp: func() error {
			return writeStatusCommand(defaultVCSRevision, defaultVCSTime, defaultVCSStatus)
		},
	})
}

func statusCommand() string {
	if runtime.GOOS == "windows" {
		return `.\\status.bat`
	}
	return "./status.sh"
}

func writeStatusCommand(revision, stamp, status string) error {
	if runtime.GOOS == "windows" {
		return os.WriteFile("status.bat", []byte(fmt.Sprintf("@echo off\r\necho STABLE_BUILD_SCM_VCS git\r\necho STABLE_BUILD_SCM_REVISION %s\r\necho STABLE_BUILD_SCM_TIME %s\r\necho STABLE_BUILD_SCM_STATUS %s\r\n", revision, stamp, status)), 0o666)
	}
	return os.WriteFile("status.sh", []byte(fmt.Sprintf("#!/bin/sh\n\necho \"STABLE_BUILD_SCM_VCS git\"\necho \"STABLE_BUILD_SCM_REVISION %s\"\necho \"STABLE_BUILD_SCM_TIME %s\"\necho \"STABLE_BUILD_SCM_STATUS %s\"\n", revision, stamp, status)), 0o755)
}

func mustWriteStatusCommand(t *testing.T, revision, stamp, status string) {
	t.Helper()
	if err := writeStatusCommand(revision, stamp, status); err != nil {
		t.Fatalf("write status command: %v", err)
	}
}

type dep struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

type withDepOutput struct {
	OK          bool              `json:"ok"`
	Path        string            `json:"path"`
	MainPath    string            `json:"main_path"`
	MainVersion string            `json:"main_version"`
	Settings    map[string]string `json:"settings"`
	Deps        []dep             `json:"deps"`
}

type stdlibOnlyOutput struct {
	OK          bool              `json:"ok"`
	Path        string            `json:"path"`
	MainPath    string            `json:"main_path"`
	MainVersion string            `json:"main_version"`
	DepCount    int               `json:"dep_count"`
	Settings    map[string]string `json:"settings"`
}

func TestReadBuildInfoDeps(t *testing.T) {
	stdout, err := bazel_testing.BazelOutput("run", "//:with_dep")
	if err != nil {
		t.Fatal(err)
	}

	var got withDepOutput
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("unmarshal output %q: %v", stdout, err)
	}
	if !got.OK {
		t.Fatalf("ReadBuildInfo returned ok=false: %+v", got)
	}
	if got.Path != "with_dep" {
		t.Fatalf("got Path %q; want %q", got.Path, "with_dep")
	}
	if got.Settings["-buildmode"] != expectedBuildMode() {
		t.Fatalf("got -buildmode %q; want %q", got.Settings["-buildmode"], expectedBuildMode())
	}
	if got.Settings["-compiler"] != "gc" {
		t.Fatalf("got -compiler %q; want gc", got.Settings["-compiler"])
	}
	if got.Settings["-trimpath"] != "true" {
		t.Fatalf("got -trimpath %q; want true", got.Settings["-trimpath"])
	}
	if got.Settings["GOOS"] != runtime.GOOS {
		t.Fatalf("got GOOS %q; want %q", got.Settings["GOOS"], runtime.GOOS)
	}
	if got.Settings["GOARCH"] != runtime.GOARCH {
		t.Fatalf("got GOARCH %q; want %q", got.Settings["GOARCH"], runtime.GOARCH)
	}
	if got.MainPath != "example.com/main" || got.MainVersion != "v1.0.0" {
		t.Fatalf("got Main %q %q; want example.com/main v1.0.0", got.MainPath, got.MainVersion)
	}
	if len(got.Deps) == 0 {
		t.Fatalf("got no deps: %+v", got)
	}

	foundCmp := false
	foundUnusedDeclaredDep := false
	for _, dep := range got.Deps {
		if dep.Path == "github.com/google/go-cmp" && dep.Version == "v0.6.0" {
			foundCmp = true
		}
		if dep.Path == "example.com/versionless" && dep.Version == "(devel)" {
			foundUnusedDeclaredDep = true
		}
	}
	if !foundCmp {
		t.Fatalf("missing github.com/google/go-cmp@v0.6.0 in %+v", got.Deps)
	}
	if !foundUnusedDeclaredDep {
		t.Fatalf("missing unused declared dependency example.com/versionless@(devel) in %+v", got.Deps)
	}
	for _, dep := range got.Deps {
		if dep.Path == "example.com/main" {
			t.Fatalf("binary's own metadata was included as a dependency: %+v", got.Deps)
		}
	}
}

func TestReadBuildInfoEmbeddedMain(t *testing.T) {
	stdout, err := bazel_testing.BazelOutput("run", "//:embedded_wrapper")
	if err != nil {
		t.Fatal(err)
	}

	var got withDepOutput
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("unmarshal output %q: %v", stdout, err)
	}
	if got.MainPath != "example.com/main" || got.MainVersion != "v1.0.0" {
		t.Fatalf("got Main %q %q; want example.com/main v1.0.0", got.MainPath, got.MainVersion)
	}
	for _, dep := range got.Deps {
		if dep.Path == "example.com/main" {
			t.Fatalf("embedded main module was included as a dependency: %+v", got.Deps)
		}
	}
}

func expectedBuildMode() string {
	switch runtime.GOOS {
	case "android", "darwin", "ios", "windows":
		return "pie"
	default:
		return "exe"
	}
}

func TestReadBuildInfoTags(t *testing.T) {
	stdout, err := bazel_testing.BazelOutput("run", "--@io_bazel_rules_go//go/config:tags=foo,bar", "//:with_dep")
	if err != nil {
		t.Fatal(err)
	}

	var got withDepOutput
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("unmarshal output %q: %v", stdout, err)
	}
	if got.Settings["-tags"] != "foo,bar" {
		t.Fatalf("got -tags %q; want foo,bar", got.Settings["-tags"])
	}
}

func TestReadBuildInfoWithoutMetadata(t *testing.T) {
	stdout, err := bazel_testing.BazelOutput("run", "//:stdlib_only")
	if err != nil {
		t.Fatal(err)
	}

	var got stdlibOnlyOutput
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("unmarshal output %q: %v", stdout, err)
	}
	if !got.OK {
		t.Fatalf("ReadBuildInfo returned ok=false: %+v", got)
	}
	if got.Path != "stdlib_only" {
		t.Fatalf("got Path %q; want %q", got.Path, "stdlib_only")
	}
	if got.MainPath != "" || got.MainVersion != "" {
		t.Fatalf("got Main %q %q; want empty", got.MainPath, got.MainVersion)
	}
	if got.DepCount != 0 {
		t.Fatalf("got %d deps; want 0", got.DepCount)
	}
	for _, key := range []string{"vcs", "vcs.revision", "vcs.time", "vcs.modified"} {
		if value := got.Settings[key]; value != "" {
			t.Fatalf("got %s=%q in unstamped build info; want empty", key, value)
		}
	}
}

func TestReadBuildInfoStampedWithoutMainModuleOmitsVCSSettings(t *testing.T) {
	stdout, err := bazel_testing.BazelOutput("run", "--stamp", "--workspace_status_command="+statusCommand(), "//:stdlib_only")
	if err != nil {
		t.Fatal(err)
	}

	var got stdlibOnlyOutput
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("unmarshal output %q: %v", stdout, err)
	}
	if !got.OK {
		t.Fatalf("ReadBuildInfo returned ok=false: %+v", got)
	}
	if got.MainPath != "" || got.MainVersion != "" {
		t.Fatalf("got Main %q %q; want empty", got.MainPath, got.MainVersion)
	}
	for _, key := range []string{"vcs", "vcs.revision", "vcs.time", "vcs.modified"} {
		if value := got.Settings[key]; value != "" {
			t.Fatalf("got %s=%q in stamped build without module metadata; want empty", key, value)
		}
	}
}

func TestGoTestBuildInfoDeps(t *testing.T) {
	if _, err := bazel_testing.BazelOutput("test", "//:with_dep_test"); err != nil {
		t.Fatal(err)
	}
}

func TestGoTestBuildInfoDepsWithCoverage(t *testing.T) {
	if _, err := bazel_testing.BazelOutput("coverage", "//:with_dep_test"); err != nil {
		t.Fatal(err)
	}
}

func TestReadBuildInfoVersionlessDep(t *testing.T) {
	stdout, err := bazel_testing.BazelOutput("run", "//:with_versionless_dep")
	if err != nil {
		t.Fatal(err)
	}

	var got withDepOutput
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("unmarshal output %q: %v", stdout, err)
	}
	if !got.OK {
		t.Fatalf("ReadBuildInfo returned ok=false: %+v", got)
	}
	if got.Path != "with_versionless_dep" {
		t.Fatalf("got Path %q; want %q", got.Path, "with_versionless_dep")
	}

	foundVersionless := false
	for _, dep := range got.Deps {
		if dep.Path == "example.com/versionless" && dep.Version == "(devel)" {
			foundVersionless = true
			break
		}
	}
	if !foundVersionless {
		t.Fatalf("missing example.com/versionless@(devel) in %+v", got.Deps)
	}
}

func TestReadBuildInfoVendoredDep(t *testing.T) {
	stdout, err := bazel_testing.BazelOutput("run", "//:with_vendored_dep")
	if err != nil {
		t.Fatal(err)
	}

	var got withDepOutput
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("unmarshal output %q: %v", stdout, err)
	}
	if !got.OK {
		t.Fatalf("ReadBuildInfo returned ok=false: %+v", got)
	}

	for _, dep := range got.Deps {
		if dep.Path == "example.com/vendored" && dep.Version == "v1.2.3" {
			return
		}
	}
	t.Fatalf("missing example.com/vendored@v1.2.3 in %+v", got.Deps)
}

func TestReadBuildInfoVCSSettingsWithExternalSource(t *testing.T) {
	stdout, err := bazel_testing.BazelOutput("run", "--stamp", "--workspace_status_command="+statusCommand(), "//:with_external_src")
	if err != nil {
		t.Fatal(err)
	}

	var got stdlibOnlyOutput
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("unmarshal output %q: %v", stdout, err)
	}
	if !got.OK {
		t.Fatalf("ReadBuildInfo returned ok=false: %+v", got)
	}
	if got.MainPath != "example.com/main" || got.MainVersion != "v1.0.0" {
		t.Fatalf("got Main %q %q; want example.com/main v1.0.0", got.MainPath, got.MainVersion)
	}
	if got.Settings["vcs"] != "git" {
		t.Fatalf("got vcs %q; want %q", got.Settings["vcs"], "git")
	}
	if got.Settings["vcs.revision"] != "abc123" {
		t.Fatalf("got vcs.revision %q; want %q", got.Settings["vcs.revision"], "abc123")
	}
}

func TestReadBuildInfoVCSSettingsAqueryUsesFilteredStableStamp(t *testing.T) {
	action, err := bazel_testing.BazelOutput("aquery", "--stamp", "--workspace_status_command="+statusCommand(), `mnemonic("GoLink", //:with_dep)`)
	if err != nil {
		t.Fatal(err)
	}

	got := string(action)
	if strings.Contains(got, "stable-status.txt") {
		t.Fatalf("unexpected stable-status stamp input in action:\n%s", got)
	}
	if !strings.Contains(got, ".vcsstamp") {
		t.Fatalf("missing filtered VCS stamp input in action:\n%s", got)
	}
}

func TestReadBuildInfoStableXDefsAqueryUsesStableStatus(t *testing.T) {
	action, err := bazel_testing.BazelOutput("aquery", "--stamp", "--workspace_status_command="+statusCommand(), `mnemonic("GoLink", //:with_stable_xdef)`)
	if err != nil {
		t.Fatal(err)
	}

	got := string(action)
	if !strings.Contains(got, "stable-status.txt") {
		t.Fatalf("missing stable-status stamp input in action:\n%s", got)
	}
	if !strings.Contains(got, ".vcsstamp") {
		t.Fatalf("missing filtered VCS stamp input in action:\n%s", got)
	}
}

func TestReadBuildInfoVCSSettingsEmbeddedExternalMain(t *testing.T) {
	stdout, err := bazel_testing.BazelOutput("run", "--stamp", "--workspace_status_command="+statusCommand(), "//:embedded_external_wrapper")
	if err != nil {
		t.Fatal(err)
	}

	var got stdlibOnlyOutput
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("unmarshal output %q: %v", stdout, err)
	}
	if !got.OK {
		t.Fatalf("ReadBuildInfo returned ok=false: %+v", got)
	}
	if got.MainPath != "example.com/externalembedded" || got.MainVersion != "v4.5.6" {
		t.Fatalf("got Main %q %q; want example.com/externalembedded v4.5.6", got.MainPath, got.MainVersion)
	}
	for _, key := range []string{"vcs", "vcs.revision", "vcs.time", "vcs.modified"} {
		if value := got.Settings[key]; value != "" {
			t.Fatalf("got %s=%q in embedded external stamped build info; want empty", key, value)
		}
	}
}

func TestReadBuildInfoVCSSettingsEmbeddedExternalMainAquery(t *testing.T) {
	action, err := bazel_testing.BazelOutput("aquery", "--stamp", "--workspace_status_command="+statusCommand(), `mnemonic("GoLink", //:embedded_external_wrapper)`)
	if err != nil {
		t.Fatal(err)
	}

	got := string(action)
	if strings.Contains(got, "stable-status.txt") {
		t.Fatalf("unexpected stable-status stamp input in action:\n%s", got)
	}
}

func TestReadBuildInfoVCSSettingsExternalMetadataMain(t *testing.T) {
	stdout, err := bazel_testing.BazelOutput("run", "--stamp", "--workspace_status_command="+statusCommand(), "//:metadata_external_wrapper")
	if err != nil {
		t.Fatal(err)
	}

	var got stdlibOnlyOutput
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("unmarshal output %q: %v", stdout, err)
	}
	if !got.OK {
		t.Fatalf("ReadBuildInfo returned ok=false: %+v", got)
	}
	if got.MainPath != "example.com/externalmetadata" || got.MainVersion != "v7.8.9" {
		t.Fatalf("got Main %q %q; want example.com/externalmetadata v7.8.9", got.MainPath, got.MainVersion)
	}
	for _, key := range []string{"vcs", "vcs.revision", "vcs.time", "vcs.modified"} {
		if value := got.Settings[key]; value != "" {
			t.Fatalf("got %s=%q in external metadata stamped build info; want empty", key, value)
		}
	}
}

func TestReadBuildInfoVCSSettingsExternalMetadataMainAquery(t *testing.T) {
	action, err := bazel_testing.BazelOutput("aquery", "--stamp", "--workspace_status_command="+statusCommand(), `mnemonic("GoLink", //:metadata_external_wrapper)`)
	if err != nil {
		t.Fatal(err)
	}

	got := string(action)
	if strings.Contains(got, "stable-status.txt") {
		t.Fatalf("unexpected stable-status stamp input in action:\n%s", got)
	}
}

func TestReadBuildInfoGoTestExternalMetadata(t *testing.T) {
	if _, err := bazel_testing.BazelOutput("test", "--stamp", "--workspace_status_command="+statusCommand(), "//:subject_external_metadata_test"); err != nil {
		t.Fatal(err)
	}
}

func TestReadBuildInfoVCSSettings(t *testing.T) {
	stdout, err := bazel_testing.BazelOutput("run", "--stamp", "--workspace_status_command="+statusCommand(), "//:with_dep")
	if err != nil {
		t.Fatal(err)
	}

	var got withDepOutput
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("unmarshal output %q: %v", stdout, err)
	}
	if !got.OK {
		t.Fatalf("ReadBuildInfo returned ok=false: %+v", got)
	}
	if got.Settings["vcs"] != "git" {
		t.Fatalf("got vcs %q; want %q", got.Settings["vcs"], "git")
	}
	if got.Settings["vcs.revision"] != "abc123" {
		t.Fatalf("got vcs.revision %q; want %q", got.Settings["vcs.revision"], "abc123")
	}
	if got.Settings["vcs.time"] != "2026-04-21T10:47:02Z" {
		t.Fatalf("got vcs.time %q; want %q", got.Settings["vcs.time"], "2026-04-21T10:47:02Z")
	}
	if got.Settings["vcs.modified"] != "true" {
		t.Fatalf("got vcs.modified %q; want %q", got.Settings["vcs.modified"], "true")
	}
}

func TestReadBuildInfoVCSSettingsCustomBinary(t *testing.T) {
	stdout, err := bazel_testing.BazelOutput("run", "--stamp", "--workspace_status_command="+statusCommand(), "//:custom_stamped")
	if err != nil {
		t.Fatal(err)
	}

	var got stdlibOnlyOutput
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("unmarshal output %q: %v", stdout, err)
	}
	if got.MainPath != "example.com/main" || got.MainVersion != "v1.0.0" {
		t.Fatalf("got Main %q %q; want example.com/main v1.0.0", got.MainPath, got.MainVersion)
	}
	if got.Settings["vcs.revision"] != defaultVCSRevision {
		t.Fatalf("got vcs.revision %q; want %q", got.Settings["vcs.revision"], defaultVCSRevision)
	}
}

func TestReadBuildInfoVCSSettingsCustomMultiOutput(t *testing.T) {
	if _, err := bazel_testing.BazelOutput("build", "--stamp", "--workspace_status_command="+statusCommand(), "//:custom_multi_output"); err != nil {
		t.Fatal(err)
	}
}

func TestReadBuildInfoVCSSettingsMainPackageNamedExternal(t *testing.T) {
	action, err := bazel_testing.BazelOutput("aquery", "--stamp", "--workspace_status_command="+statusCommand(), `mnemonic("GoLink", //external/tool:bin)`)
	if err != nil {
		t.Fatal(err)
	}

	got := string(action)
	if !strings.Contains(got, "-main_workspace=true") {
		t.Fatalf("missing -main_workspace=true in action:\n%s", got)
	}
}

func TestReadBuildInfoVCSSettingsRefreshOnStableStampChange(t *testing.T) {
	t.Cleanup(func() {
		mustWriteStatusCommand(t, defaultVCSRevision, defaultVCSTime, defaultVCSStatus)
	})
	run := func() withDepOutput {
		stdout, err := bazel_testing.BazelOutput("run", "--stamp", "--workspace_status_command="+statusCommand(), "//:with_dep")
		if err != nil {
			t.Fatal(err)
		}
		var got withDepOutput
		if err := json.Unmarshal(stdout, &got); err != nil {
			t.Fatalf("unmarshal output %q: %v", stdout, err)
		}
		if !got.OK {
			t.Fatalf("ReadBuildInfo returned ok=false: %+v", got)
		}
		return got
	}

	mustWriteStatusCommand(t, defaultVCSRevision, defaultVCSTime, defaultVCSStatus)
	first := run()
	if first.Settings["vcs.revision"] != defaultVCSRevision {
		t.Fatalf("got first vcs.revision %q; want %q", first.Settings["vcs.revision"], defaultVCSRevision)
	}
	if first.Settings["vcs.modified"] != "true" {
		t.Fatalf("got first vcs.modified %q; want %q", first.Settings["vcs.modified"], "true")
	}

	mustWriteStatusCommand(t, "def456", "2026-04-22T08:15:00+07:00", "Clean")
	second := run()
	if second.Settings["vcs.revision"] != "def456" {
		t.Fatalf("got second vcs.revision %q; want %q", second.Settings["vcs.revision"], "def456")
	}
	if second.Settings["vcs.time"] != "2026-04-22T01:15:00Z" {
		t.Fatalf("got second vcs.time %q; want %q", second.Settings["vcs.time"], "2026-04-22T01:15:00Z")
	}
	if second.Settings["vcs.modified"] != "false" {
		t.Fatalf("got second vcs.modified %q; want %q", second.Settings["vcs.modified"], "false")
	}
}

func TestReadBuildInfoVCSSettingsExternalBinary(t *testing.T) {
	// External repository binaries should not inherit vcs.* from the caller's
	// workspace status command.
	stdout, err := bazel_testing.BazelOutput("run", "--stamp", "--workspace_status_command="+statusCommand(), "@com_example_external_binary//:tool")
	if err != nil {
		t.Fatal(err)
	}

	var got stdlibOnlyOutput
	if err := json.Unmarshal(stdout, &got); err != nil {
		t.Fatalf("unmarshal output %q: %v", stdout, err)
	}
	if !got.OK {
		t.Fatalf("ReadBuildInfo returned ok=false: %+v", got)
	}
	if got.MainPath != "example.com/externaltool" || got.MainVersion != "v9.9.9" {
		t.Fatalf("got Main %q %q; want example.com/externaltool v9.9.9", got.MainPath, got.MainVersion)
	}
	for _, key := range []string{"vcs", "vcs.revision", "vcs.time", "vcs.modified"} {
		if value := got.Settings[key]; value != "" {
			t.Fatalf("got %s=%q in external stamped build info; want empty", key, value)
		}
	}
}

func TestReadBuildInfoVCSSettingsExternalBinaryAquery(t *testing.T) {
	action, err := bazel_testing.BazelOutput("aquery", "--stamp", "--workspace_status_command="+statusCommand(), `mnemonic("GoLink", @com_example_external_binary//:tool)`)
	if err != nil {
		t.Fatal(err)
	}

	got := string(action)
	if !strings.Contains(got, "-main_workspace") || !strings.Contains(got, "false") {
		t.Fatalf("missing -main_workspace=false in action:\n%s", got)
	}
	if strings.Contains(got, "stable-status.txt") {
		t.Fatalf("unexpected stable-status stamp input in action:\n%s", got)
	}
}

func TestReadBuildInfoVCSSettingsGoTestBinary(t *testing.T) {
	if _, err := bazel_testing.BazelOutput("test", "--stamp", "--workspace_status_command="+statusCommand(), "//:subject_stamped_test"); err != nil {
		t.Fatal(err)
	}
}
