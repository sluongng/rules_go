## Defines and stamping

In order to provide build time information to go code without data files, we
support the concept of stamping.

Stamping asks the linker to substitute the value of a global variable with a
string determined at link time. Stamping only happens when linking a binary, not
when compiling a package. This means that changing a value results only in
re-linking, not re-compilation and thus does not cause cascading changes.

Link values are set in the `x_defs` attribute of any Go rule. This is a
map of string to string, where keys are the names of variables to substitute,
and values are the string to use. Keys may be names of variables in the package
being compiled, or they may be fully qualified names of variables in another
package.

These mappings are collected up across the entire transitive dependencies of a
binary. This means you can set a value using `x_defs` in a
`go_library`, and any binary that links that library will be stamped with that
value. You can also override stamp values from libraries using `x_defs`
on the `go_binary` rule if needed. The `--[no]stamp` option controls whether
stamping of workspace variables is enabled.

The values of the `x_defs` dictionary are subject to
[location expansion](https://bazel.build/reference/be/make-variables#predefined_label_variables).

**Example**

Suppose we have a small library that contains the current version.

``` go
package version

var Version = "redacted"
```

We can set the version in the `go_library` rule for this library.

``` bzl
go_library(
    name = "version",
    srcs = ["version.go"],
    importpath = "example.com/repo/version",
    x_defs = {"Version": "0.9"},
)
```

Binaries that depend on this library may also set this value.

``` bzl
go_binary(
    name = "cmd",
    srcs = ["main.go"],
    deps = ["//version"],
    x_defs = {"example.com/repo/version.Version": "0.9"},
)
```

### Stamping with the workspace status script

You can use values produced by the workspace status command in your link stamp.
To use this functionality, write a script that prints key-value pairs, separated
by spaces, one per line. For example:

``` bash
#!/usr/bin/env bash

echo STABLE_GIT_COMMIT $(git rev-parse HEAD)
```

***Note:*** stamping with keys that bazel designates as "stable" will trigger a
re-link when any stable key changes. Currently, in bazel, stable keys are
`BUILD_EMBED_LABEL`, `BUILD_USER`, `BUILD_HOST` and keys whose names start with
`STABLE_`. Stamping only with keys that are not stable keys will not trigger a
relink.

You can reference these in `x_defs` using curly braces.

``` bzl
go_binary(
    name = "cmd",
    srcs = ["main.go"],
    deps = ["//version"],
    x_defs = {"example.com/repo/version.Version": "{STABLE_GIT_COMMIT}"},
)
```

You can build using the status script using the `--workspace_status_command`
argument on the command line:

``` bash
$ bazel build --stamp --workspace_status_command=./status.sh //:cmd
```

### VCS build info

When `go_binary` and `go_test` targets are linked with stamping enabled,
`rules_go` also maps a small stable workspace status convention into Go build
info. These settings are visible through `runtime/debug.ReadBuildInfo` and
`go version -m`. They are only emitted when the linked target and its recorded
main module both come from the main workspace; stamped binaries built from
external repositories, targets without main module metadata, and local wrappers
around external main modules omit `vcs.*` settings. Main-module provenance
follows the metadata source that supplied `Main.Path`, including embedded
libraries and `package_metadata` / `applicable_licenses`.

| Workspace status key | Go build info setting | Notes |
| --- | --- | --- |
| `STABLE_BUILD_SCM_VCS` | `vcs` | Required and must be non-empty before any `vcs.*` settings are emitted. |
| `STABLE_BUILD_SCM_REVISION` | `vcs.revision` | Optional when non-empty. |
| `STABLE_BUILD_SCM_TIME` | `vcs.time` | Optional; must be `RFC3339Nano` and is normalized to UTC. |
| `STABLE_BUILD_SCM_STATUS` | `vcs.modified` | Optional; `Clean` becomes `false`, `Modified` becomes `true`. |

Invalid or missing optional values are ignored individually.

``` bash
#!/usr/bin/env bash

echo "STABLE_BUILD_SCM_VCS git"
echo "STABLE_BUILD_SCM_REVISION $(git rev-parse HEAD)"
echo "STABLE_BUILD_SCM_TIME $(git show -s --format=%cI HEAD)"
if git diff-index --quiet HEAD --; then
  echo "STABLE_BUILD_SCM_STATUS Clean"
else
  echo "STABLE_BUILD_SCM_STATUS Modified"
fi
```
