// Performs file conversion
package converterservice

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	_ "github.com/lib/pq"
	"github.com/reggiemcdonald/grpc-audio-converter/pb"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	ffmpeg     = "ffmpeg"
	formatFlag = "-f"
	inputFlag  = "-i"
)

type FileConverterConfiguration struct {
	s3endpoint string
	bucketName string
	region     string
	dbUser     string
	dbPass     string
	isDev      bool
}

type FileConverter struct {
	s3 *s3.S3
	uploader *s3manager.Uploader
	bucketName string
	db *FileConverterData
	isDev bool
}

func NewFileConverter(config FileConverterConfiguration) *FileConverter {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(config.region),
		Endpoint: aws.String(config.s3endpoint),
		S3ForcePathStyle: aws.Bool(true),
	}))
	db := NewFileConverterData(config.dbUser, config.dbPass)
	return &FileConverter{
		s3: s3.New(sess),
		uploader: s3manager.NewUploader(sess),
		bucketName: config.bucketName,
		db: db,
		isDev: config.isDev,
	}
}

/*
 * Formats the encoding to its string representation
 */
func encodingToString(req *pb.ConvertFileRequest) (string, string) {
	sourceEncoding := req.GetSourceEncoding().String()
	destEncoding   := req.GetDestEncoding().String()
	return sourceEncoding, destEncoding
}

/*
 * Creates a signed GET url for the converted file
 */
func (f *FileConverter) signedUrl(id string) (string, error) {
	req, _ := f.s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(f.bucketName),
		Key: aws.String(id),
	})
	url, err := req.Presign(24 * time.Hour)
	if f.isDev {
		dockerNetworkName := "s3_local"
		localhost := "localhost"
		url = strings.Replace(url, dockerNetworkName, localhost, 1)
	}
	if err != nil {
		return "", err
	}
	return url, nil
}

/*
 * Downloads a file at the request source URL and streams it to ffmpeg for conversion
 * to the requested encoding
 */
func (f *FileConverter) ConvertFile(req *pb.ConvertFileRequest, id string) {
	if _, err := f.db.NewRequest(id); err != nil {
		log.Printf("could not create new job, encounterd %v", err)
		return
	}
	sourceUrl := req.SourceUrl
	sourceEncoding, destEncoding := encodingToString(req)
	tmpFile := fmt.Sprintf("/tmp/%s", id)
	cmd := exec.Command(ffmpeg,
		formatFlag,
		sourceEncoding,
		inputFlag,
		sourceUrl,
		formatFlag,
		destEncoding,
		tmpFile,
	)
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Printf("failed to start conversion due to: %v", err)
		return
	}
	if err := cmd.Wait(); err != nil {
		log.Printf("conversion failed, ecnountered %v", err)
		if _, err := f.db.FailConversion(id); err != nil {
			log.Printf("Failed to update job status, encountered %v", err)
		}
		return
	}
	file, err := os.Open(tmpFile)
	if err != nil {
		log.Fatalf("error preserving tmp file %v", err)
	}
	if _, err := f.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(f.bucketName),
		Key:    aws.String(id),
		Body:   file,
		ContentType: aws.String(fmt.Sprintf("audio/%s", destEncoding)),
	}); err != nil {
		log.Printf("failed to upload converted audio to S3, ecnountered %v", err)
	}
	if err := os.Remove(tmpFile); err != nil {
		panic(err)
	}
	url, err := f.signedUrl(id)
	if err != nil {
		log.Printf("Failed to generate presigned URL for ID %s", id)
		if _, err := f.db.FailConversion(id); err != nil {
			log.Printf("failed to update job status, encountered %v", err)
		}
		return
	}
	if _, err := f.db.CompleteConversion(id, url); err != nil {
		log.Printf("failed to update DB for ID %s, encountered %v", id, err)
	} else {
		log.Printf("%s successfully converted", id)
	}
}
