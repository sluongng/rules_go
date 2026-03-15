/* Copyright 2026 The Bazel Authors. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import "strings"

func normalizeGoVersion(goVersion string) string {
	if goVersion == "" {
		return ""
	}
	if !strings.HasPrefix(goVersion, "go") {
		goVersion = "go" + goVersion
	}
	return normalizeGoVersionForTypes(goVersion)
}

func trimGoPatchVersion(goVersion string) string {
	prefix := ""
	if strings.HasPrefix(goVersion, "go") {
		prefix = "go"
		goVersion = strings.TrimPrefix(goVersion, "go")
	}
	parts := strings.SplitN(goVersion, ".", 3)
	if len(parts) < 3 {
		return prefix + goVersion
	}
	return prefix + parts[0] + "." + parts[1]
}
