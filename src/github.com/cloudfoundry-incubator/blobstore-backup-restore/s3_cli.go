package blobstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

func NewS3CLI(awsCliPath, endpoint, regionName, accessKeyId, accessKeySecret string) S3Cli {
	return S3Cli{
		awsCliPath:      awsCliPath,
		endpoint:        endpoint,
		regionName:      regionName,
		accessKeyId:     accessKeyId,
		accessKeySecret: accessKeySecret,
	}
}

type S3Cli struct {
	awsCliPath      string
	endpoint        string
	regionName      string
	accessKeyId     string
	accessKeySecret string
}

func (s3Cli S3Cli) ListObjects(s string) ([]string, error) {
	output, err := s3Cli.run("list-objects", "--bucket", s)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(string(output)) == "" {
		return []string{}, nil
	}

	response := S3ListResponse{}
	err = json.Unmarshal(output, &response)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, object := range response.Contents {
		files = append(files, object.Key)
	}

	return files, nil
}

func (s3Cli S3Cli) ListObjectVersions(s string) ([]byte, error) {
	return s3Cli.run("list-object-versions", "--bucket", s)
}

func (s3Cli S3Cli) run(args ...string) ([]byte, error) {
	outputBuffer := new(bytes.Buffer)
	errorBuffer := new(bytes.Buffer)

	var baseArgs []string
	if s3Cli.endpoint != "" {
		baseArgs = []string{"--output", "json", "--region", s3Cli.regionName, "--endpoint", s3Cli.endpoint, "s3api"}
	} else {
		baseArgs = []string{"--output", "json", "--region", s3Cli.regionName, "s3api"}
	}

	awsCmd := exec.Command(s3Cli.awsCliPath, append(baseArgs, args...)...)
	awsCmd.Env = append(awsCmd.Env, "AWS_ACCESS_KEY_ID="+s3Cli.accessKeyId)
	awsCmd.Env = append(awsCmd.Env, "AWS_SECRET_ACCESS_KEY="+s3Cli.accessKeySecret)
	awsCmd.Stdout = outputBuffer
	awsCmd.Stderr = errorBuffer

	err := awsCmd.Run()
	if err != nil {
		return nil, errors.New(errorBuffer.String())
	}

	return outputBuffer.Bytes(), nil
}
