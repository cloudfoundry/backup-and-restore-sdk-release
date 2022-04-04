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

import (
	"fmt"
	"strconv"
	"time"

	"reflect"
)

//go:generate go-bindata -pkg binmock -o packaged_client.go client/
type Mock struct {
	Path                string
	identifier          string
	currentMappingIndex int
	failHandler         FailHandler

	mappings    []*InvocationStub
	invocations []Invocation
}

// The type of the function that will be invoked when an assertion fails. Compatible with the ginkgo fail handler (`ginkgo.Fail`)
type FailHandler func(message string, callerSkip ...int)

// Creates a new binary mock
func NewBinMock(failHandler FailHandler) *Mock {
	server := getCurrentServer()

	identifier := strconv.FormatInt(time.Now().UnixNano(), 10)
	binaryPath, err := buildBinary(identifier, server.listener.Addr().String())
	if err != nil {
		failHandler(fmt.Sprintf("cant build binary %v", err))
	}

	mock := &Mock{identifier: identifier, Path: binaryPath, failHandler: failHandler}

	server.monitor(mock)
	return mock
}

func (mock *Mock) invoke(args, env, stdin []string) (int, string, string) {
	if mock.currentMappingIndex >= len(mock.mappings) {
		mock.failHandler(fmt.Sprintf("Too many calls to the mock! Last call with %v", args))
		return 1, "", ""
	}
	currentMapping := mock.mappings[mock.currentMappingIndex]
	mock.currentMappingIndex = mock.currentMappingIndex + 1
	if currentMapping.expectedArgs != nil && !reflect.DeepEqual(currentMapping.expectedArgs, args) {
		mock.failHandler(fmt.Sprintf("Expected %v to equal %v", args, currentMapping.expectedArgs))
		return 1, "", ""
	}
	mock.invocations = append(mock.invocations, newInvocation(args, env, stdin))
	return currentMapping.exitCode, currentMapping.stdout, currentMapping.stderr
}

// Sets up a stub for a possible invocation of the mock, accepting any arguments
func (mock *Mock) WhenCalled() *InvocationStub {
	return mock.createMapping(&InvocationStub{})
}

// Sets up a stub for a possible invocation of the mock, with specific arguments
// If args don't match the actual arguments to the mock then it fails
func (mock *Mock) WhenCalledWith(args ...string) *InvocationStub {
	invocation := &InvocationStub{}
	invocation.expectedArgs = args
	return mock.createMapping(invocation)
}

func (mock *Mock) createMapping(mapping *InvocationStub) *InvocationStub {
	mock.mappings = append(mock.mappings, mapping)
	return mapping
}

// Invocations returns the list of invocations of the mock till now
func (mock *Mock) Invocations() []Invocation {
	return mock.invocations
}

// Resets the mapping and invocations to the mock
func (mock *Mock) Reset() {
	mock.mappings = []*InvocationStub{}
	mock.invocations = []Invocation{}
	mock.currentMappingIndex = 0
}
