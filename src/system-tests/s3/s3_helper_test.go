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

package s3_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	. "github.com/onsi/gomega"
)

type PutResponse struct {
	VersionId string
}

type ListResponse struct {
	Contents []ListResponseEntry
}

type ListResponseEntry struct {
	Key string
}

type VersionsResponse struct {
	Versions      []VersionItem
	DeleteMarkers []VersionItem
}

type VersionItem struct {
	Key       string
	VersionId string
}

func GetFileContentsFromBucket(region, bucket, key string) string {
	outputBuffer := runAwsCommandOnBucketSuccessfully(
		"--region", region,
		"s3",
		"cp",
		fmt.Sprintf("s3://%s/%s", bucket, key),
		"-")

	return outputBuffer.String()
}

func ListFilesFromBucket(region, bucket string) []string {
	outputBuffer := runAwsCommandOnBucketSuccessfully(
		"--region", region,
		"s3api",
		"list-objects",
		"--bucket", bucket)

	var response ListResponse
	_ = json.Unmarshal(outputBuffer.Bytes(), &response)

	var keys []string
	for _, entry := range response.Contents {
		keys = append(keys, entry.Key)
	}

	return keys
}

func UploadTimestampedFileToBucket(region, bucket, prefix, body string) string {
	fileName := prefix + "_" + strconv.FormatInt(time.Now().Unix(), 10)
	WriteFileInBucket(region, bucket, fileName, body)
	return fileName
}

func runAwsCommandOnBucketSuccessfully(args ...string) *bytes.Buffer {
	outputBuffer := new(bytes.Buffer)
	errorBuffer := new(bytes.Buffer)

	cmdArgs := append([]string{"--profile", assumedRoleProfileName}, args...)
	awsCmd := exec.Command("aws", cmdArgs...)
	awsCmd.Stdout = outputBuffer
	awsCmd.Stderr = errorBuffer

	err := awsCmd.Run()
	Expect(err).ToNot(HaveOccurred(), errorBuffer.String())

	return outputBuffer
}

func WriteFileInBucket(region, bucket, key, body string) {
	bodyFile, _ := os.CreateTemp("", "")
	_, _ = bodyFile.WriteString(body)
	bodyFile.Close()

	runAwsCommandOnBucketSuccessfully(
		"--region", region,
		"s3api",
		"put-object",
		"--bucket", bucket,
		"--key", key,
		"--body", bodyFile.Name())
}

func deleteVersionFromBucket(region, bucket, key, versionId string) string {
	outputBuffer := runAwsCommandOnBucketSuccessfully(
		"--region", region,
		"s3api",
		"delete-object",
		"--bucket", bucket,
		"--key", key,
		"--version-id", versionId)

	var response PutResponse
	_ = json.Unmarshal(outputBuffer.Bytes(), &response)

	return response.VersionId
}

func DeleteAllVersionsFromBucket(region, bucket string) {
	outputBuffer := runAwsCommandOnBucketSuccessfully(
		"--region", region,
		"s3api",
		"list-object-versions",
		"--bucket", bucket)

	var response VersionsResponse
	_ = json.Unmarshal(outputBuffer.Bytes(), &response)

	for _, version := range response.Versions {
		deleteVersionFromBucket(region, bucket, version.Key, version.VersionId)
	}

	for _, version := range response.DeleteMarkers {
		deleteVersionFromBucket(region, bucket, version.Key, version.VersionId)
	}
}

func DeleteFileFromBucket(region, bucket, key string) string {
	outputBuffer := runAwsCommandOnBucketSuccessfully(
		"--region", region,
		"s3api",
		"delete-object",
		"--bucket", bucket,
		"--key", key)

	var response PutResponse
	_ = json.Unmarshal(outputBuffer.Bytes(), &response)

	return response.VersionId
}

func DeleteAllFilesFromBucket(region, bucket string) {
	runAwsCommandOnBucketSuccessfully(
		"--region", region,
		"s3",
		"rm",
		"--recursive",
		fmt.Sprintf("s3://%s", bucket),
	)
}
