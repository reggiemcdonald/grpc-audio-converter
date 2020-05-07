package converterservice

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"os"
	"strings"
	"time"
)

type S3Service interface {
	Upload(id string, encoding string, file *os.File) error
	SignedUrl(id string) (string, error)
}

type s3Service struct {
	s3 *s3.S3
	uploader *s3manager.Uploader
	bucket string
}

type localS3Service struct {
	s3 *s3.S3
	uploader *s3manager.Uploader
	bucket string
}

func NewS3Service(region string, endpoint string, bucket string) S3Service {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
		Endpoint: aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(true),
	}))
	return &s3Service{
		s3: s3.New(sess),
		uploader: s3manager.NewUploader(sess),
		bucket: bucket,
	}
}

func NewLocalS3Service(region string, endpoint string, bucket string) S3Service {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
		Endpoint: aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(true),
	}))
	return &localS3Service{
		s3: s3.New(sess),
		uploader: s3manager.NewUploader(sess),
		bucket: bucket,
	}
}

func upload(bucket string, id string, encoding string, file *os.File, uploader *s3manager.Uploader) error {
	if _, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(id),
		Body:   file,
		ContentType: aws.String(fmt.Sprintf("audio/%s", encoding)),
	}); err != nil {
		return err
	}
	return nil
}

func signedUrl(bucket string, id string, s *s3.S3) (string, error) {
	req, _ := s.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key: aws.String(id),
	})
	return req.Presign(24 * time.Hour)
}

func (s *s3Service) Upload(id string, encoding string, file *os.File) error {
	return upload(s.bucket, id, encoding, file, s.uploader)
}

func (s *s3Service) SignedUrl(id string) (string, error) {
	return signedUrl(s.bucket, id, s.s3)
}

func (l *localS3Service) Upload(id string, encoding string, file *os.File) error {
	return upload(l.bucket, id, encoding, file, l.uploader)
}

func (l *localS3Service) SignedUrl(id string) (string, error) {
	url, err := signedUrl(l.bucket, id, l.s3)
	if err != nil {
		return url, err
	}
	dockerNetworkName := "s3_local"
	localhost := "localhost"
	url = strings.Replace(url, dockerNetworkName, localhost, 1)
	return url, err
}

