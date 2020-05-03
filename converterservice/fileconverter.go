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
	ffmpeg      = "ffmpeg"
	formatFlag  = "-f"
	inputFlag   = "-i"
	mapFlag     = "-map"
	audioStream = "0:0"
)

type FileConverterService interface {
	ConvertFile(request *pb.ConvertFileRequest, id string)
}

type FileConverterConfiguration struct {
	s3endpoint string
	bucketName string
	region     string
	db         FileConverterDataService
	isDev      bool
}

type FileConverter struct {
	s3 *s3.S3
	uploader *s3manager.Uploader
	bucketName string
	db FileConverterDataService
	isDev bool
}

type conversionAttributes struct {
	id               string
	sourceEncoding   string
	sourceUrl        string
	destEncoding     string
	includeExtension bool
	tmpFile          string
}

func NewFileConverter(config FileConverterConfiguration) *FileConverter {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(config.region),
		Endpoint: aws.String(config.s3endpoint),
		S3ForcePathStyle: aws.Bool(true),
	}))
	return &FileConverter{
		s3: s3.New(sess),
		uploader: s3manager.NewUploader(sess),
		bucketName: config.bucketName,
		db: config.db,
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
 * Creates the file path for the temp file created during the conversion process.
 * Includes file extension when includeExtension is set to true
 */
func newTempFilePath(id string, destEncoding string, includeExtension bool) string {
	if includeExtension {
		return fmt.Sprintf("/tmp/%s.%s", id, destEncoding)
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
		job.sourceEncoding,
		inputFlag,
		job.sourceUrl,
		mapFlag,
		audioStream,
		formatFlag,
		job.destEncoding,
		job.tmpFile)
}

/*
 * Creates a command object for conversions to MP4.
 * Note: MPEG-4 is the container type, and M4A specifies audio only
 * so we force the extension to be the audio type
 */
func commandForMP4(job *conversionAttributes) *exec.Cmd {
	job.tmpFile = newTempFilePath(job.id, "m4a", job.includeExtension)
	return commandForDestEncoding(job)
}

/*
 * Creates a command object for codecs that do not require special circumstances
 */
func defaultCommand(job *conversionAttributes) *exec.Cmd {
	job.tmpFile = newTempFilePath(job.id, job.destEncoding, job.includeExtension)
	return commandForDestEncoding(job)
}

func selectCommand(job *conversionAttributes) (cmd *exec.Cmd){
	fmt.Printf("THE FIEL FORMAT IS %s", job.destEncoding)
	switch job.destEncoding {
	case "MP4":
		cmd = commandForMP4(job)
	default:
		cmd = defaultCommand(job)
	}
	return cmd
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
	job := &conversionAttributes{
		id: id,
		sourceEncoding: sourceEncoding,
		sourceUrl: sourceUrl,
		destEncoding: destEncoding,
		// TODO: add this to the request
		includeExtension: false,
	}
	cmd := selectCommand(job)
	fmt.Printf("THE FILE PATH IS %s", job.tmpFile)
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
	file, err := os.Open(job.tmpFile)
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
	if err := os.Remove(job.tmpFile); err != nil {
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
