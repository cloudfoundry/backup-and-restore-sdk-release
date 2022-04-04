// Copyright (C) 2017-Present Pivotal Software, Inc. All rights reserved.
//
// This program and the accompanying materials are made available under
// the terms of the under the Apache License, Version 2.0 (the "License”);
// you may not use this file except in compliance with the License.
//
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.

package binmock

import "strings"

// Invocation represents an invocation of the mock
type Invocation struct {
	args  []string
	env   map[string]string
	stdin []string
}

func newInvocation(args, env, stdin []string) Invocation {
	return Invocation{
		args:  args,
		env:   parseEnv(env),
		stdin: stdin,
	}
}

// Args represents the arguments passed to the mock when it was invoked
func (invocation Invocation) Args() []string {
	return invocation.args
}

// Env represents the environment at the time of invocation
func (invocation Invocation) Env() map[string]string {
	return invocation.env
}

// Stdin represents the standard input steam received by the mock as a slice of lines
func (invocation Invocation) Stdin() []string {
	return invocation.stdin
}

func parseEnv(envVars []string) map[string]string {
	parsedVars := map[string]string{}

	for _, v := range envVars {
		parts := strings.Split(v, "=")

		parsedVars[parts[0]] = parts[1]
	}
	return parsedVars
}
