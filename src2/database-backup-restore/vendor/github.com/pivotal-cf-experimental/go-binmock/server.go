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
	"encoding/json"
	"net"
	"net/http"
)

type server struct {
	mocks map[string]*Mock
	*http.Server
	listener net.Listener
}

var currentServer *server

func getCurrentServer() *server {
	if currentServer == nil {
		currentServer = &server{
			mocks: map[string]*Mock{},
		}
		currentServer.start()
	}
	return currentServer
}

func (server *server) start() {
	server.Server = &http.Server{Addr: ":0", Handler: http.HandlerFunc(server.serve)}
	server.listener, _ = net.Listen("tcp", "127.0.0.1:0")
	go server.Server.Serve(server.listener)
}

type invocationRequest struct {
	Id    string
	Args  []string
	Env   []string
	Stdin []string
}

type invocationResponse struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func newInvocationResponse(exitCode int, stdout, stderr string) invocationResponse {
	return invocationResponse{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
	}
}

func (server *server) serve(resp http.ResponseWriter, req *http.Request) {
	invocationRequest := invocationRequest{}
	json.NewDecoder(req.Body).Decode(&invocationRequest)
	currentMock := server.mocks[invocationRequest.Id]
	invocationResponse := newInvocationResponse(currentMock.invoke(invocationRequest.Args, invocationRequest.Env, invocationRequest.Stdin))
	json.NewEncoder(resp).Encode(invocationResponse)
}

func (server *server) monitor(mock *Mock) {
	server.mocks[mock.identifier] = mock
}
