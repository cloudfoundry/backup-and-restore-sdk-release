package blobstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

//go:generate counterfeiter -o fakes/fake_bucket.go . Bucket
type Bucket interface {
	Name() string
	RegionName() string
	Versions() ([]Version, error)
	PutVersions(regionName, bucketName string, versions []LatestVersion) error
}

type S3Bucket struct {
	awsCliPath string
	name       string
	regionName string
	accessKey  S3AccessKey
	endpoint   string
}

type S3AccessKey struct {
	Id     string
	Secret string
}

func NewS3Bucket(awsCliPath, name, region, endpoint string, accessKey S3AccessKey) S3Bucket {
	return S3Bucket{
		awsCliPath: awsCliPath,
		name:       name,
		regionName: region,
		accessKey:  accessKey,
		endpoint:   endpoint,
	}
}

func (b S3Bucket) Name() string {
	return b.name
}

func (b S3Bucket) RegionName() string {
	return b.regionName
}

func (b S3Bucket) Versions() ([]Version, error) {
	s3Cli := NewS3CLI(b.awsCliPath, b.endpoint, b.regionName, b.accessKey.Id, b.accessKey.Secret)

	output, err := s3Cli.ListObjectVersions(b.name)

	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(string(output)) == "" {
		return []Version{}, nil
	}

	response := S3ListVersionsResponse{}
	err = json.Unmarshal(output, &response)
	if err != nil {
		return nil, err
	}

	return response.Versions, nil
}

func (b S3Bucket) PutVersions(regionName, bucketName string, versions []LatestVersion) error {
	var err error

	s3Cli := NewS3CLI(b.awsCliPath, b.endpoint, b.regionName, b.accessKey.Id, b.accessKey.Secret)
	s3API, err := NewS3API(b.endpoint, b.regionName, b.accessKey.Id, b.accessKey.Secret)

	if err != nil {
		return err
	}

	for _, version := range versions {
		err = s3API.PutVersion(bucketName, version.BlobKey, version.Id)
		if err != nil {
			return err
		}
	}

	files, err := s3Cli.ListObjects(bucketName)
	if err != nil {
		return err
	}

	for _, file := range files {
		included := versionsIncludeFile(file, versions)
		if !included {
			err = s3API.DeleteObject(bucketName, file)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type S3ListVersionsResponse struct {
	Versions []Version
}

type Version struct {
	Key      string
	Id       string `json:"VersionId"`
	IsLatest bool
}

type S3ListResponse struct {
	Contents []Object
}

type Object struct {
	Key string
}

func versionsIncludeFile(file string, versions []LatestVersion) bool {
	for _, version := range versions {
		if version.BlobKey == file {
			return true
		}
	}

	return false
}
