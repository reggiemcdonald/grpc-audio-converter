// Performs file conversion
package converterservice

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	_ "github.com/lib/pq"
	encodings "github.com/reggiemcdonald/grpc-audio-converter/converterservice/enums"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	ffmpeg      = "ffmpeg"
	formatFlag  = "-f"
	inputFlag   = "-i"
	mapFlag     = "-map"
	audioStream = "0:0"
)

type FileConverterService interface {
	ConvertFile(request *FileConversionRequest)
}

type FileConverterConfiguration struct {
	S3endpoint string
	BucketName string
	Region     string
	Db         FileConverterDataRepository
	IsDev      bool
}

type FileConverter struct {
	s3         *s3.S3
	uploader   *s3manager.Uploader
	bucketName string
	db         FileConverterDataRepository
	isDev      bool
}

type conversionAttributes struct {
	request          *FileConversionRequest
	tmpFile          string
}

func NewFileConverter(config *FileConverterConfiguration) *FileConverter {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(config.Region),
		Endpoint: aws.String(config.S3endpoint),
		S3ForcePathStyle: aws.Bool(true),
	}))
	return &FileConverter{
		s3: s3.New(sess),
		uploader: s3manager.NewUploader(sess),
		bucketName: config.BucketName,
		db: config.Db,
		isDev: config.IsDev,
	}
}

/*
 * Creates the file path for the temp file created during the conversion process.
 * Includes file extension when includeExtension is set to true
 */
func newTempFilePath(id string, destEncoding string, includeExtension bool) string {
	if includeExtension {
		return fmt.Sprintf("/tmp/%s.%s", id, strings.ToLower(destEncoding))
	}
	return fmt.Sprintf("/tmp/%s", id)
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
 * Returns a pointer to the command object
 */
func commandForDestEncoding(job *conversionAttributes) *exec.Cmd {
	return exec.Command(ffmpeg,
		formatFlag,
		job.request.SourceEncoding.Name(),
		inputFlag,
		job.request.SourceUrl,
		mapFlag,
		audioStream,
		formatFlag,
		job.request.DestEncoding.Name(),
		job.tmpFile)
}

/*
 * Creates a command object for conversions to MP4.
 * Note: MPEG-4 is the container type, and M4A specifies audio only
 * so we force the extension to be the audio type
 */
func commandForMP4(job *conversionAttributes) *exec.Cmd {
	job.tmpFile = newTempFilePath(job.request.Id, "m4a", job.request.IncludeExtension)
	return commandForDestEncoding(job)
}

/*
 * Creates a command object for codecs that do not require special circumstances
 */
func defaultCommand(job *conversionAttributes) *exec.Cmd {
	job.tmpFile = newTempFilePath(job.request.Id, job.request.DestEncoding.Name(), job.request.IncludeExtension)
	return commandForDestEncoding(job)
}

/*
 * Selects the appropriate command to be created
 */
func selectCommand(job *conversionAttributes) (cmd *exec.Cmd){
	switch job.request.DestEncoding {
	case encodings.MP4:
		cmd = commandForMP4(job)
	default:
		cmd = defaultCommand(job)
	}
	return cmd
}

/*
 * Downloads a file at the request source URL and streams it to ffmpeg for conversion
 * to the requested name
 */
func (f *FileConverter) ConvertFile(req *FileConversionRequest) {
	id := req.Id
	if _, err := f.db.NewRequest(id); err != nil {
		log.Printf("could not create new job, encounterd %v", err)
		return
	}
	job := &conversionAttributes{
		request: req,
	}
	cmd := selectCommand(job)
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Printf("failed to start conversion due to: %v", err)
		return
	}
	if err := cmd.Wait(); err != nil {
		log.Printf("conversion failed, ecnountered %v", err)
		if _, err := f.db.FailConversion(id); err != nil {
			log.Printf("Failed to update job Status, encountered %v", err)
		}
		return
	}
	file, err := os.Open(job.tmpFile)
	if err != nil {
		log.Fatalf("error preserving tmp file %v", err)
	}
	if _, err := f.uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(f.bucketName),
		Key:    aws.String(id),
		Body:   file,
		ContentType: aws.String(fmt.Sprintf("audio/%s", req.DestEncoding.Name())),
	}); err != nil {
		log.Printf("failed to upload converted audio to S3, ecnountered %v", err)
	}
	if err := os.Remove(job.tmpFile); err != nil {
		panic(err)
	}
	url, err := f.signedUrl(id)
	if err != nil {
		log.Printf("Failed to generate presigned URL for Id %s", id)
		if _, err := f.db.FailConversion(id); err != nil {
			log.Printf("failed to update job Status, encountered %v", err)
		}
		return
	}
	if _, err := f.db.CompleteConversion(id, url); err != nil {
		log.Printf("failed to update DB for Id %s, encountered %v", id, err)
	} else {
		log.Printf("%s successfully converted", id)
	}
}
