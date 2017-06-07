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

// tags takes a list of go source files and does tag filtering on them.
package main

import (
	"fmt"
	"go/build"
	"log"
	"os"
	"strings"
)

type FilterFlags struct {
	Cgo    bool   `help:"Sets whether cgo-using files are allowed to pass the filter."`
	Quiet  bool   `help:"Don't print filenames. Return code will be 0 if any files pass the filter"`
	Tags   string `help:"Only pass through files that match these tags"`
	Output string `help:"If set, write the matching files to this file, instead of stdout"`
}

func run(args []string) error {
	// Prepare our flags
	flags := FilterFlags{}
	files, err := parseFlags("filter_tags", &flags, "filters filenames using go build tag logic", args)
	if err != nil {
		return err
	}
	// filter our input file list
	bctx := build.Default
	bctx.CgoEnabled = flags.Cgo
	bctx.BuildTags = strings.Split(flags.Tags, ",")
	filtered, err := filterFiles(bctx, files)
	if err != nil {
		return err
	}
	// if we are in quiet mode, just vary our exit condition based on the results
	if flags.Quiet {
		if len(filtered) == 0 {
			os.Exit(1)
		}
		return nil
	}
	// print the outputs if we need not
	to := os.Stdout
	if flags.Output != "" {
		f, err := os.Create(flags.Output)
		if err != nil {
			return err
		}
		defer f.Close()
		to = f
	}
	for _, filename := range filtered {
		fmt.Fprintln(to, filename)
	}

	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		log.Fatal(err)
	}
}
