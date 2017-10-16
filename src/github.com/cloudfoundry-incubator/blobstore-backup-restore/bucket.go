package blobstore

import (
	"bytes"
	"encoding/json"
	"errors"
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
	RegionName() string
	Versions() ([]Version, error)
}

type S3Bucket struct {
	awsCliPath string
	name       string
	regionName string
	accessKey  S3AccessKey
}

func NewS3Bucket(awsCliPath, name, region string, accessKey S3AccessKey) S3Bucket {
	return S3Bucket{awsCliPath: awsCliPath, name: name, regionName: region, accessKey: accessKey}
}

func (b S3Bucket) Name() string {
	return b.name
}

func (b S3Bucket) RegionName() string {
	return b.regionName
}

func (b S3Bucket) Versions() ([]Version, error) {
	outputBuffer := new(bytes.Buffer)
	errorBuffer := new(bytes.Buffer)

	awsCmd := exec.Command(b.awsCliPath,
		"--output", "json",
		"--region", b.regionName,
		"s3api",
		"list-object-versions",
		"--bucket", b.name)
	awsCmd.Env = append(awsCmd.Env, "AWS_ACCESS_KEY_ID="+b.accessKey.Id)
	awsCmd.Env = append(awsCmd.Env, "AWS_SECRET_ACCESS_KEY="+b.accessKey.Secret)
	awsCmd.Stdout = outputBuffer
	awsCmd.Stderr = errorBuffer

	err := awsCmd.Run()
	if err != nil {
		return nil, errors.New(errorBuffer.String())
	}

	response := S3ListVersionsResponse{}
	err = json.Unmarshal(outputBuffer.Bytes(), &response)
	if err != nil {
		return nil, err
	}

	return response.Versions, nil
}

type S3AccessKey struct {
	Id     string
	Secret string
}

type S3ListVersionsResponse struct {
	Versions []Version
}
