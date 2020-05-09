package mocks

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

type S3FileUploaderMock struct {
	bucket   string
	endpoint string
	region   string
	Success  bool
}

type LocalFileUploaderMock struct {
	bucket   string
	endpoint string
	region   string
	Success  bool
}

func NewMockS3FileUploader(region string, endpoint string, bucket string) *S3FileUploaderMock {
	return &S3FileUploaderMock{
		bucket: bucket,
		endpoint: endpoint,
		region: region,
		Success: true,
	}
}

func NewMockLocalFileUploader(region string, endpoint string, bucket string) *LocalFileUploaderMock {
	return &LocalFileUploaderMock{
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

func (m *S3FileUploaderMock) Upload(id string, encoding string, file *os.File) error {
	if m.Success {
		return Upload(id, encoding, file)
	}
	return errors.New(fmt.Sprintf("failed to upload %s", id))
}

func (m *S3FileUploaderMock) SignedUrl(id string) (string, error) {
	if m.Success {
		return SignedUrl(m.region, m.endpoint, m.bucket, id), nil
	}
	return "", errors.New(fmt.Sprintf("failed to get signed URL for %s", id))
}

func (m *LocalFileUploaderMock) Upload(id string, encoding string, file *os.File) error {
	if m.Success {
		return Upload(id, encoding, file)
	}
	return errors.New(fmt.Sprintf("failed to upload %s", id))
}

func (m *LocalFileUploaderMock) SignedUrl(id string) (string, error) {
	if m.Success {
		url := SignedUrl(m.region, m.endpoint, m.bucket, id)
		dockerNetworkName := "s3_local"
		localhost := "localhost"
		url = strings.Replace(url, dockerNetworkName, localhost, 1)
		return url, nil
	}
	return "", errors.New(fmt.Sprintf("failed to get signed URL for %s", id))
}
