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

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func vcsStamp(args []string) error {
	args, _, err := expandParamsFiles(args)
	if err != nil {
		return err
	}

	flags := flag.NewFlagSet("vcsstamp", flag.ExitOnError)
	input := flags.String("in", "", "The stable workspace status file to filter.")
	output := flags.String("out", "", "The filtered VCS stamp file to write.")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *input == "" {
		return fmt.Errorf("-in must be set")
	}
	if *output == "" {
		return fmt.Errorf("-out must be set")
	}

	stampMap, err := readStampMap([]string{*input})
	if err != nil {
		return err
	}

	entries := buildInfoVCSEntries(stampMap)
	var builder strings.Builder
	for _, entry := range entries {
		builder.WriteString(entry.key)
		builder.WriteByte(' ')
		builder.WriteString(entry.value)
		builder.WriteByte('\n')
	}
	return os.WriteFile(*output, []byte(builder.String()), 0o666)
}
