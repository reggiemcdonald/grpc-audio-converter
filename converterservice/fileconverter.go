package converterservice

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/reggiemcdonald/grpc-audio-converter/pb"
	"log"
	"os"
	"os/exec"
)

const (
	FFMPEG      = "ffmpeg"
	FORMAT_FLAG = "-f"
	INPUT_FLAG  = "-i"
	STDIN_PIPE  = "pipe:0"
)

type FileConverter struct {
	uploader *s3manager.Uploader
}

func NewFileConverter(s3Endpoint string) *FileConverter {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("REGION")),
		Endpoint: aws.String(os.Getenv("S3_ENDPOINT")),
		Credentials: credentials.NewStaticCredentialsFromCreds(credentials.Value{
			AccessKeyID: os.Getenv("ACCESS_KEY"),
			SecretAccessKey: os.Getenv("SECRET_ACCESS_KEY"),
		}),
		S3ForcePathStyle: aws.Bool(true),
	}))
	f := FileConverter{
		uploader: s3manager.NewUploader(sess),
	}
	return &f
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
 * Downloads a file at the request source URL and streams it to FFMPEG for conversion
 * to the requested encoding
 */
func (f *FileConverter) ConvertFile(req *pb.ConvertFileRequest, id string) {
	sourceUrl := req.SourceUrl
	destinationBucket := os.Getenv("BUCKET_NAME")
	sourceEncoding, destEncoding := encodingToString(req)
	tmpFile := fmt.Sprintf("/tmp/%s", id)
	cmd := exec.Command(FFMPEG,
		FORMAT_FLAG,
		sourceEncoding,
		INPUT_FLAG,
		sourceUrl,
		FORMAT_FLAG,
		destEncoding,
		tmpFile,
	)
	cmd.Stderr = os.Stderr
	cmd.Start()
	cmd.Wait()
	file, err := os.Open(tmpFile)
	if err != nil {
		log.Fatalf("error preserving tmp file %v", err)
	}
	if _, err := f.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(destinationBucket),
		Key:    aws.String(id),
		Body:   file,
	}); err != nil {
		log.Printf("error: %v", err)
		log.Fatal("Failed to send to s3")
	}
	if err := os.Remove(tmpFile); err != nil {
		panic(err)
	}
	log.Printf("sent %s to s3", id)
}