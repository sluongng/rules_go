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
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
)

func readStampMap(files []string) (map[string]string, error) {
	stampMap := map[string]string{}
	for _, stampFile := range files {
		stampBuf, err := os.ReadFile(stampFile)
		if err != nil {
			return nil, fmt.Errorf("failed reading stamp file %s: %v", stampFile, err)
		}
		scanner := bufio.NewScanner(bytes.NewReader(stampBuf))
		for scanner.Scan() {
			line := strings.SplitN(scanner.Text(), " ", 2)
			switch len(line) {
			case 0:
				// Nothing to do here.
			case 1:
				stampMap[line[0]] = ""
			case 2:
				stampMap[line[0]] = line[1]
			}
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed scanning stamp file %s: %v", stampFile, err)
		}
	}
	return stampMap, nil
}
