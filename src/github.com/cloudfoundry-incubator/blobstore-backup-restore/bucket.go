package blobstore

import (
	"bytes"
	"encoding/json"
	"os/exec"
)

type Version struct {
	Key      string
	Id       string `json:"VersionId"`
	IsLatest bool
}

//go:generate counterfeiter -o fakes/fake_bucket.go . Bucket
type Bucket interface {
	Name() string
	Versions() []Version // TODO: error!
}

type S3Bucket struct {
	name      string
	region    string
	accessKey s3AccessKey
}

func NewS3Bucket(name, region string, accessKey s3AccessKey) S3Bucket {
	return S3Bucket{name: name, region: region, accessKey: accessKey}
}

func (b S3Bucket) Name() string {
	return b.name
}

func (b S3Bucket) Versions() []Version {
	outputBuffer := new(bytes.Buffer)
	awsCmd := exec.Command("aws",
		"--output", "json",
		"--region", b.region,
		"s3api",
		"list-object-versions",
		"--bucket", b.name)
	awsCmd.Env = append(awsCmd.Env, "AWS_ACCESS_KEY_ID="+b.accessKey.Id)
	awsCmd.Env = append(awsCmd.Env, "AWS_SECRET_ACCESS_KEY="+b.accessKey.Secret)
	awsCmd.Stdout = outputBuffer
	awsCmd.Run()

	response := s3ListVersionsResponse{}
	json.Unmarshal(outputBuffer.Bytes(), &response)

	return response.Versions
}

type s3AccessKey struct {
	Id     string
	Secret string
}

type s3ListVersionsResponse struct {
	Versions []Version
}
