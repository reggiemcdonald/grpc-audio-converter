package mocks

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

type S3ServiceMock struct {
	bucket   string
	endpoint string
	region   string
	Success  bool
}

type LocalS3ServiceMock struct {
	bucket   string
	endpoint string
	region   string
	Success  bool
}

func NewMockS3Service(region string, endpoint string, bucket string) *S3ServiceMock {
	return &S3ServiceMock{
		bucket: bucket,
		endpoint: endpoint,
		region: region,
		Success: true,
	}
}

func NewMockLocalS3Service(region string, endpoint string, bucket string) *LocalS3ServiceMock {
	return &LocalS3ServiceMock{
		bucket: bucket,
		endpoint: endpoint,
		region: region,
		Success: true,
	}
}

func Upload(id string, encoding string, file *os.File) error {
	log.Printf("uploading id %s...\n", id)
	return nil
}

func SignedUrl(region string, endpoint string, bucket string, id string) string {
	return fmt.Sprintf("http://%s.%s/%s/%s", region, endpoint, bucket, id)
}

func (m *S3ServiceMock) Upload(id string, encoding string, file *os.File) error {
	if m.Success {
		return Upload(id, encoding, file)
	}
	return errors.New(fmt.Sprintf("failed to upload %s", id))
}

func (m *S3ServiceMock) SignedUrl(id string) (string, error) {
	if m.Success {
		return SignedUrl(m.region, m.endpoint, m.bucket, id), nil
	}
	return "", errors.New(fmt.Sprintf("failed to get signed URL for %s", id))
}

func (m *LocalS3ServiceMock) Upload(id string, encoding string, file *os.File) error {
	if m.Success {
		return Upload(id, encoding, file)
	}
	return errors.New(fmt.Sprintf("failed to upload %s", id))
}

func (m *LocalS3ServiceMock) SignedUrl(id string) (string, error) {
	if m.Success {
		url := SignedUrl(m.region, m.endpoint, m.bucket, id)
		dockerNetworkName := "s3_local"
		localhost := "localhost"
		url = strings.Replace(url, dockerNetworkName, localhost, 1)
		return url, nil
	}
	return "", errors.New(fmt.Sprintf("failed to get signed URL for %s", id))
}
