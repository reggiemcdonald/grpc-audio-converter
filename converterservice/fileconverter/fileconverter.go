// Performs file conversion
package fileconverter

import (
	"fmt"
	_ "github.com/lib/pq"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice/db"
	"log"
	"os"
	"strings"
)

const (
	ffmpeg      = "ffmpeg"
	formatFlag  = "-f"
	inputFlag   = "-i"
	mapFlag     = "-map"
	audioStream = "0:0"
)

type Converter interface {
	ConvertFile(request *FileConversionRequest)
}

type ConverterImplementation struct {
	Db                db.FileConverterRepository
	ExecutableFactory ExecutableFactory
	S3service         FileUploader
}

type FileConverter struct {
	s3Service         FileUploader
	db                db.FileConverterRepository
	executableFactory ExecutableFactory
}

type ConversionAttributes struct {
	Request *FileConversionRequest
	TmpFile string
}

// An init function for the file converter
func New(config *ConverterImplementation) *FileConverter {
	s3Service := config.S3service
	factory := config.ExecutableFactory
	if factory == nil {
		factory = newDefaultExecutableFactory()
	}
	return &FileConverter{
		s3Service: s3Service,
		db: config.Db,
		executableFactory: factory,
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
 * Returns a pointer to the command object
 */
func commandForDestEncoding(job *ConversionAttributes) Executable {
	return newDefaultExecutable(ffmpeg,
		formatFlag,
		job.Request.SourceEncoding.Name(),
		inputFlag,
		job.Request.SourceUrl,
		mapFlag,
		audioStream,
		formatFlag,
		job.Request.DestEncoding.Name(),
		job.TmpFile)
}

/*
 * Downloads a file at the Request source URL and streams it to ffmpeg for conversion
 * to the requested name
 */
func (f *FileConverter) ConvertFile(req *FileConversionRequest) {
	id := req.Id
	if _, err := f.db.StartConversion(id); err != nil {
		log.Printf("failure updating job status, encounterd %v", err)
		return
	}
	job := &ConversionAttributes{
		Request: req,
	}
	cmd := f.executableFactory.Build(job)
	cmd.SetStderr(os.Stderr)
	if err := cmd.Start(); err != nil {
		log.Printf("failed to start conversion due to: %v", err)
		if _, err := f.db.FailConversion(id); err != nil {
			log.Printf("Failed to update job status, encountered %v", err)
		}
		return
	}
	if err := cmd.Wait(); err != nil {
		log.Printf("conversion failed, ecnountered %v", err)
		if _, err := f.db.FailConversion(id); err != nil {
			log.Printf("Failed to update job Status, encountered %v", err)
		}
		return
	}
	file, err := os.Open(job.TmpFile)
	if err != nil {
		log.Fatalf("error preserving tmp file %v", err)
	}
	if err := f.s3Service.Upload(id, req.DestEncoding.Name(), file); err != nil {
		log.Printf("failed to upload converted audio to S3, ecnountered %v", err)
	}
	if err := os.Remove(job.TmpFile); err != nil {
		panic(err)
	}
	url, err := f.s3Service.SignedUrl(id)
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
