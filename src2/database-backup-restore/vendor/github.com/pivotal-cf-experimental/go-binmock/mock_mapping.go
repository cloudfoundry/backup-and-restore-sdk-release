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

// InvocationStub offers a fluid API to set up the behaviour on invocation of the binary mock
type InvocationStub struct {
	expectedArgs []string

	exitCode int
	stdout   string
	stderr   string
}

// WillPrintToStdOut sets up what the mock will print to standard out on invocation
func (stub *InvocationStub) WillPrintToStdOut(out string) *InvocationStub {
	stub.stdout = out
	return stub
}

// WillPrintToStdErr sets up what the mock will print to standard error on invocation
func (stub *InvocationStub) WillPrintToStdErr(err string) *InvocationStub {
	stub.stderr = err
	return stub
}

// WillExitWith sets up the exit code of the mock invocation
func (stub *InvocationStub) WillExitWith(exitCode int) *InvocationStub {
	stub.exitCode = exitCode
	return stub
}
