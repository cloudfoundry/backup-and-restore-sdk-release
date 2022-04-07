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
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

func buildBinary(identifier, serverUrl string) (string, error) {
	clientPath, err := getSourceFile()
	if err != nil {
		return "", fmt.Errorf("cant extract client source %v", err)
	}

	binaryPath, err := doBuild(clientPath, "-ldflags", "-X main.serverUrl="+serverUrl+" -X main.identifier="+identifier)

	if err != nil {
		return "", fmt.Errorf("can't build binary %v", err)
	}

	err = os.Remove(clientPath)
	if err != nil {
		return "", fmt.Errorf("can't remove client source %v", err)
	}
	return binaryPath, nil
}

func getSourceFile() (string, error) {
	data, err := Asset("client/main.go")
	if err != nil {
		return "", err
	}

	tempFile, err := ioutil.TempFile("", "go-bindata-client")
	if err != nil {
		return "", err
	}

	_, err = tempFile.Write(data)
	if err != nil {
		return "", err
	}
	err = tempFile.Close()
	if err != nil {
		return "", err
	}

	sourceFilePath := tempFile.Name() + ".go"
	err = os.Rename(tempFile.Name(), sourceFilePath)
	if err != nil {
		return "", err
	}
	return sourceFilePath, nil
}

func doBuild(packagePath string, args ...string) (compiledPath string, err error) {
	tmpDir, err := ioutil.TempDir("", "bin_mock")
	if err != nil {
		return "", err
	}

	executable := filepath.Join(tmpDir, path.Base(packagePath))
	cmdArgs := append([]string{"build"}, args...)
	cmdArgs = append(cmdArgs, "-o", executable, packagePath)

	build := exec.Command("go", cmdArgs...)

	output, err := build.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed to build %s:\n\nError:\n%s\n\nOutput:\n%s", packagePath, err, string(output))
	}

	return executable, nil
}
