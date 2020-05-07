// Tests FileConverter
package converterservice_test

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice"
	encodings "github.com/reggiemcdonald/grpc-audio-converter/converterservice/enums"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice/mocks"
	"github.com/reggiemcdonald/grpc-audio-converter/pb"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

const (
	testS3Endpoint = "test-endpoint"
	testBucketName = "test-bucket-name"
	testRegion     = "test-Region"
)

func defaultTestingConfiguration() *converterservice.FileConverterConfiguration {
	return &converterservice.FileConverterConfiguration{
		Db: mocks.NewMockFileConverterDb(),
		ExecutableFactory: mocks.NewMockExecutableFactory(),
	}
}

func TestNewFileConverter(t *testing.T) {
	config := defaultTestingConfiguration()
	t.Run("online s3 service", func (t *testing.T) {
		config.S3service = mocks.NewMockS3Service(testRegion, testS3Endpoint, testBucketName)
		fileconverter := converterservice.NewFileConverter(config)
		assert.NotNil(t, fileconverter, "file converter should not be nil")
	})
	t.Run("locla S3 service", func (t *testing.T) {
		config.S3service = mocks.NewMockLocalS3Service(testRegion, testS3Endpoint, testBucketName)
		fileconverter := converterservice.NewFileConverter(config)
		assert.NotNil(t, fileconverter, "file converter should not be nil")
	})
}

func TestConvertFile_Success(t *testing.T) {
	req := &converterservice.FileConversionRequest{
		Id: uuid.New().String(),
		SourceUrl: "some-source-url",
		SourceEncoding: encodings.FLAC,
		DestEncoding: encodings.MP3,
	}
	repo := mocks.NewMockFileConverterDb()
	executableFactory := mocks.NewMockExecutableFactory()
	s3Service := mocks.NewMockS3Service(testRegion, testS3Endpoint, testBucketName)
	config := &converterservice.FileConverterConfiguration{
		Db: repo,
		ExecutableFactory: executableFactory,
		S3service: s3Service,
	}
	fileConverter := converterservice.NewFileConverter(config)
	success, err := config.Db.NewRequest(req.Id)
	assert.True(t, success, "new request should have been successful")
	assert.Nil(t, err, "should not have errored adding to the repo")
	fileConverter.ConvertFile(req)
	convertedJob, err := repo.GetConversion(req.Id)
	assert.Nil(t, err, "err should be nil")
	assert.NotNil(t, convertedJob, "converted job should not be nil")
	assert.Equal(t, req.Id, convertedJob.Id, "Id should be the same")
	assert.Equal(t, pb.ConvertFileQueryResponse_COMPLETED.String(), convertedJob.Status, "status should be complete")
	assert.Equal(t,
		fmt.Sprintf("http://%s.%s/%s/%s", testRegion, testS3Endpoint, testBucketName, req.Id),
		convertedJob.CurrUrl, "should have the correct presigned URL")
	assert.GreaterOrEqual(t, time.Now().Unix(), convertedJob.LastUpdated.Unix(), "should have been updated previously")
	file, err := os.Open(executableFactory.Data[req.Id].Job.TmpFile)
	assert.Nil(t, file, "there should be no file once completed conversion")
	assert.NotNil(t, err, "there should have been an error opening the file")
}

func TestConvertFile_FailedCmd(t *testing.T) {
	req := &converterservice.FileConversionRequest{
		Id: uuid.New().String(),
		SourceUrl: "some-source-url",
		SourceEncoding: encodings.FLAC,
		DestEncoding: encodings.MP3,
	}
	repo := mocks.NewMockFileConverterDb()
	executableFactory := mocks.NewMockExecutableFactory()
	s3Service := mocks.NewMockS3Service(testRegion, testS3Endpoint, testBucketName)
	config := &converterservice.FileConverterConfiguration{
		Db: repo,
		ExecutableFactory: executableFactory,
		S3service: s3Service,
	}
	fileConverter := converterservice.NewFileConverter(config)
	executableFactory.Success = false
	fileConverter.ConvertFile(req)
	convertedJob, err := repo.GetConversion(req.Id)
	assert.Nil(t, err, "err should be nil")
	assert.NotNil(t, convertedJob, "converted job should not be nil")
	assert.Equal(t, req.Id, convertedJob.Id, "Id should be the same")
	assert.Equal(t, pb.ConvertFileQueryResponse_FAILED.String(), convertedJob.Status, "should have a failed status")
	assert.Equal(t, "NONE", convertedJob.CurrUrl, "should have no presigned URL")
	assert.GreaterOrEqual(t, time.Now().Unix(), convertedJob.LastUpdated.Unix(), "should have been updated previously")
	file, err := os.Open(executableFactory.Data[req.Id].Job.TmpFile)
	assert.Nil(t, file, "there should be no file in tmp after error")
	assert.NotNil(t, err, "there should have been an error opening the file")
}

func TestConvertFile_FailedRepo(t *testing.T) {
	req := &converterservice.FileConversionRequest{
		Id: uuid.New().String(),
		SourceUrl: "some-source-url",
		SourceEncoding: encodings.FLAC,
		DestEncoding: encodings.MP3,
	}
	repo := mocks.NewMockFileConverterDb()
	executableFactory := mocks.NewMockExecutableFactory()
	s3Service := mocks.NewMockS3Service(testRegion, testS3Endpoint, testBucketName)
	config := &converterservice.FileConverterConfiguration{
		Db: repo,
		ExecutableFactory: executableFactory,
		S3service: s3Service,
	}
	fileConverter := converterservice.NewFileConverter(config)
	repo.Success = false
	fileConverter.ConvertFile(req)
	convertedJob, err := repo.GetConversion(req.Id)
	assert.Nil(t, convertedJob, "should be no entry")
	assert.NotNil(t, err, "should have errored")
	assert.Nil(t, executableFactory.Data[req.Id], "should be no job")
	file, err := os.Open(fmt.Sprintf("/tmp/%s", req.Id))
	assert.Nil(t, file, "should be no file created")
	assert.NotNil(t, err, "should not have an error")
}

func TestConvertFile_FailedS3(t *testing.T) {
	req := &converterservice.FileConversionRequest{
		Id: uuid.New().String(),
		SourceUrl: "some-source-url",
		SourceEncoding: encodings.FLAC,
		DestEncoding: encodings.MP3,
	}
	repo := mocks.NewMockFileConverterDb()
	executableFactory := mocks.NewMockExecutableFactory()
	s3Service := mocks.NewMockS3Service(testRegion, testS3Endpoint, testBucketName)
	config := &converterservice.FileConverterConfiguration{
		Db: repo,
		ExecutableFactory: executableFactory,
		S3service: s3Service,
	}
	fileConverter := converterservice.NewFileConverter(config)
	s3Service.Success = false
	fileConverter.ConvertFile(req)
	convertedJob, err := repo.GetConversion(req.Id)
	assert.NotNil(t, convertedJob, "should have a job")
	assert.Nil(t, err, "should not have errored")
	assert.Equal(t, req.Id, convertedJob.Id, "should have the same ID")
	assert.Equal(t, pb.ConvertFileQueryResponse_FAILED.String(), convertedJob.Status, "should have a failed status")
	assert.Equal(t, "NONE", convertedJob.CurrUrl, "should have no presigned URL")
	assert.GreaterOrEqual(t, time.Now().Unix(), convertedJob.LastUpdated.Unix(), "should have recent timestamp")
}

