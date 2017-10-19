package blobstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
)

//go:generate counterfeiter -o fakes/fake_bucket.go . Bucket
type Bucket interface {
	Identifier() string
	Name() string
	RegionName() string
	Versions() ([]Version, error)
}

type S3Bucket struct {
	awsCliPath string
	identifier string
	name       string
	regionName string
	accessKey  S3AccessKey
}

type S3AccessKey struct {
	Id     string
	Secret string
}

func NewS3Bucket(awsCliPath, identifier, name, region string, accessKey S3AccessKey) S3Bucket {
	return S3Bucket{
		awsCliPath: awsCliPath,
		identifier: identifier,
		name:       name,
		regionName: region,
		accessKey:  accessKey,
	}
}

func (b S3Bucket) Identifier() string {
	return b.identifier
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

type S3ListVersionsResponse struct {
	Versions []Version
}

type Version struct {
	Key      string
	Id       string `json:"VersionId"`
	IsLatest bool
}
