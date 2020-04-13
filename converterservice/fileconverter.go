package converterservice

import (
	"bufio"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/reggiemcdonald/grpc-audio-converter/pb"
	"log"
	"net/http"
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
	s3 *s3.S3
	uploader *s3manager.Uploader
}

func NewFileConverter(s3Endpoint string) *FileConverter {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
		Endpoint: aws.String(os.Getenv("S3_ENDPOINT")),
		Credentials: credentials.NewStaticCredentialsFromCreds(credentials.Value{
			AccessKeyID: aws.StringValue(aws.String("abc")),
			SecretAccessKey: aws.StringValue(aws.String("123")),
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
 * TODO: Handle errors
 */
func (f *FileConverter) ConvertFile(req *pb.ConvertFileRequest, id string) {
	sourceUrl := req.SourceUrl
	destinationBucket := os.Getenv("BUCKET_NAME")
	sourceEncoding, destEncoding := encodingToString(req)
	tmpFile := fmt.Sprintf("/tmp/%s", id)
	getFile, err := http.Get(sourceUrl)
	if err != nil {
		panic(err)
	}
	defer getFile.Body.Close()
	if err != nil {
		panic(err)
	}
	cmd := exec.Command(FFMPEG,
		FORMAT_FLAG,
		sourceEncoding,
		INPUT_FLAG,
		STDIN_PIPE,
		FORMAT_FLAG,
		destEncoding,
		tmpFile,
	)
	fileReader := bufio.NewReader(getFile.Body)
	cmd.Stdin = fileReader
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
	log.Println("here")
	cmd.Wait()
	log.Println("done")

}