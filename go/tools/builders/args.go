// Copyright 2017 The Bazel Authors. All rights reserved.
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

// asm builds a single .s file with "go tool asm". It is invoked by the
// Go rules as an action.
package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"strings"
)

// expandArgs finds any arguments starting with @ and replaces them with
// the contents of the filename following the @.
// Each line of the file becomes a separate argument preserving order.
func expandArgs(args []string) ([]string, error) {
	result := []string{}
	for _, s := range args {
		if !strings.HasPrefix(s, "@") {
			result = append(result, s)
			continue
		}
		// We have a response file, so read it in now
		content, err := ioutil.ReadFile(s)
		if err != nil {
			return nil, err
		}
		// Create a new Scanner for the file.
		for scanner := bufio.NewScanner(bytes.NewReader(content)); scanner.Scan(); {
			result = append(result, scanner.Text())
		}
	}
	return result, nil
}
