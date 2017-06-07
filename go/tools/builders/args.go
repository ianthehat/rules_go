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
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"time"
)

func parseFlags(name string, flags interface{}, help string, args []string) ([]string, error) {
	// Process the args
	args, err := expandArgs(args)
	if err != nil {
		return args, err
	}
	// build a flag set that represents the struct
	set := bindFlags(name, flags, help)
	// parse the input with the flag set
	if err := set.Parse(args); err != nil {
		return args, err
	}
	return set.Args(), nil
}

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
		content, err := ioutil.ReadFile(s[1:])
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

// multiFlag allows repeated string flags to be collected into a slice
type multiFlag struct {
	values *[]string
}

func (m multiFlag) String() string {
	if len(*m.values) == 0 {
		return ""
	}
	return fmt.Sprint(*m.values)
}

func (m multiFlag) Set(v string) error {
	(*m.values) = append(*m.values, v)
	return nil
}

// bindFlags uses reflection to bind struct members to flag values.
func bindFlags(name string, value interface{}, help string) *flag.FlagSet {
	set := flag.NewFlagSet(name, flag.ExitOnError)
	set.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", name)
		fmt.Fprintln(os.Stderr, help)
		set.PrintDefaults()
	}
	bindFlagEntry(set, name, value, help)
	return set
}

// bindFlagEntry binds a sinle field entry to flags.
// The field may be a struct of fields, or one of commonly understood types.
// Unexported/unsettable fields, or any unknown types are ignored.
func bindFlagEntry(set *flag.FlagSet, name string, value interface{}, help string) {
	switch val := value.(type) {
	case *bool:
		set.BoolVar(val, name, *val, help)
	case *int:
		set.IntVar(val, name, *val, help)
	case *int64:
		set.Int64Var(val, name, *val, help)
	case *uint:
		set.UintVar(val, name, *val, help)
	case *uint64:
		set.Uint64Var(val, name, *val, help)
	case *float64:
		set.Float64Var(val, name, *val, help)
	case *string:
		set.StringVar(val, name, *val, help)
	case *time.Duration:
		set.DurationVar(val, name, *val, help)
	case *[]string:
		set.Var(multiFlag{val}, name, help)
	case flag.Value:
		set.Var(val, name, help)
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
			return
		}
		e := rv.Elem()
		t := e.Type()
		for i := 0; i < e.NumField(); i++ {
			tf := t.Field(i)
			field := e.Field(i)
			if !field.CanSet() {
				continue // probably an unexported field
			}
			tags := tf.Tag
			fname := strings.ToLower(tf.Name)
			if explicitName := tags.Get("flag"); explicitName != "" {
				fname = explicitName
			}
			bindFlagEntry(set, fname, field.Addr().Interface(), tags.Get("help"))
		}
	}
}

func emitFlags(value interface{}) []string {
	result = []string{}
	switch val := value.(type) {
	case *bool:
		if *val {
			result = append(result, "-"+name)
		}
		set.BoolVar(val, name, *val, help)
	case *int:
		set.IntVar(val, name, *val, help)
	case *int64:
		set.Int64Var(val, name, *val, help)
	case *uint:
		set.UintVar(val, name, *val, help)
	case *uint64:
		set.Uint64Var(val, name, *val, help)
	case *float64:
		set.Float64Var(val, name, *val, help)
	case *string:
		set.StringVar(val, name, *val, help)
	case *time.Duration:
		set.DurationVar(val, name, *val, help)
	case *[]string:
		set.Var(multiFlag{val}, name, help)
	case flag.Value:
		set.Var(val, name, help)
	default:
		rv := reflect.ValueOf(value)
		if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
			return
		}
		e := rv.Elem()
		t := e.Type()
		for i := 0; i < e.NumField(); i++ {
			tf := t.Field(i)
			field := e.Field(i)
			if !field.CanSet() {
				continue // probably an unexported field
			}
			tags := tf.Tag
			fname := strings.ToLower(tf.Name)
			if explicitName := tags.Get("flag"); explicitName != "" {
				fname = explicitName
			}
			bindFlagEntry(set, fname, field.Addr().Interface(), tags.Get("help"))
		}
	}
}
