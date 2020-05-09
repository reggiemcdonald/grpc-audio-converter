// Tests the ConverterServer
package converterservice_test

import (
	"context"
	"fmt"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice/mocks"
	"github.com/reggiemcdonald/grpc-audio-converter/pb"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	testRegion = "test-region"
	testS3Endpoint = "test-endpoint"
	testBucketName = "test-bucket-name"
)

var testGrpcRequest = &pb.ConvertFileRequest{
	SourceUrl: "test-url",
	SourceEncoding: pb.Encoding_MP3,
	DestEncoding: pb.Encoding_WAV,
}

type testServerConfiguration struct {
	Port int
	S3service *mocks.S3FileUploaderMock
	Db *mocks.MockFileConverterRepo
	ExecutableFactory *mocks.MockExecutableFactory
}

func testingConfiguration() *testServerConfiguration {
	port := 3000
	s3Service := mocks.NewMockS3FileUploader(testRegion, testS3Endpoint, testBucketName)
	db := mocks.NewMockFileConverterRepo()
	return &testServerConfiguration{
		Port: port,
		S3service: s3Service,
		Db: db,
		ExecutableFactory: mocks.NewMockExecutableFactory(),
	}
}

func toServerConfiguration(testConfig *testServerConfiguration) *converterservice.ConverterServerConfig {
	return &converterservice.ConverterServerConfig{
		Db: testConfig.Db,
		ExecutableFactory: testConfig.ExecutableFactory,
		Port: testConfig.Port,
		S3service: testConfig.S3service,
	}
}

func TestNewWithConfiguration(t *testing.T) {
	config := toServerConfiguration(testingConfiguration())
	t.Run("config.ExecutableFactory=nil", func(t *testing.T) {
		newConfig := &converterservice.ConverterServerConfig{
			Port: config.Port,
			S3service: config.S3service,
			Db: config.Db,
		}
		server := converterservice.NewWithConfiguration(newConfig)
		assert.NotNil(t, server, "server should not be nil")
	})
	t.Run("config.ExecutableFactory=mockExecutableFactory", func(t *testing.T) {
		server := converterservice.NewWithConfiguration(config)
		assert.NotNil(t, server, "server should not be nil")
	})
}

func TestConverterServer_ConvertFile_Success(t *testing.T) {
	config := testingConfiguration()
	server := converterservice.NewWithConfiguration(toServerConfiguration(config))
	res, err := server.ConvertFile(context.TODO(), testGrpcRequest)
	assert.Nil(t, err, "should not have errored")
	assert.NotNil(t, res, "response should not be nil")
	assert.NotNil(t, res.Id, "response should have an ID")
	assert.True(t, res.Accepted, "request should have been accepted")
	job, err := config.Db.GetConversion(res.Id)
	assert.Nil(t, err, "there should not be an error with getting result")
	assert.NotNil(t, job, "job should not be nil")
	assert.Equal(t, res.Id, job.Id, "job should be the right job")
}

func TestConverterServer_ConvertFile_Fail(t *testing.T) {
	config := testingConfiguration()
	server := converterservice.NewWithConfiguration(toServerConfiguration(config))
	config.Db.Success = false
	res, err := server.ConvertFile(context.TODO(), testGrpcRequest)
	assert.NotNil(t, err, "should have errored")
	assert.Nil(t, res, "result should be nil")
}

func TestConverterServer_ConvertFile_SameEncoding(t *testing.T) {
	config := testingConfiguration()
	server := converterservice.NewWithConfiguration(toServerConfiguration(config))
	res, err := server.ConvertFile(context.TODO(), &pb.ConvertFileRequest{
		SourceUrl: testGrpcRequest.SourceUrl,
		SourceEncoding: pb.Encoding_WAV,
		DestEncoding: pb.Encoding_WAV,
	})
	assert.Nil(t, res, "response should be nil")
	assert.NotNil(t, err, "should have encountered an error")
}

func TestConverterServer_ConvertFile_MissingSource(t *testing.T) {
	config := testingConfiguration()
	server := converterservice.NewWithConfiguration(toServerConfiguration(config))
	res, err := server.ConvertFile(context.TODO(), &pb.ConvertFileRequest{
		SourceEncoding: pb.Encoding_WAV,
		DestEncoding: pb.Encoding_MP4,
	})
	assert.Nil(t, res, "response should be nil")
	assert.NotNil(t, err, "should have encountered an error")
}

func TestConverterServer_ConvertFileQuery_Success(t *testing.T) {
	// timeout strategy from https://stackoverflow.com/questions/24929790/how-to-set-the-go-timeout-flag-on-go-test
	timeout := time.After(3 * time.Second)
	done := make(chan bool)
	config := testingConfiguration()
	server := converterservice.NewWithConfiguration(toServerConfiguration(config))
	res, err := server.ConvertFile(context.TODO(), testGrpcRequest)
	assert.Nil(t, err, "should not have errored")
	assert.NotNil(t, res, "response should not be nil")
	job, err := config.Db.GetConversion(res.Id)
	go func() {
		for job.Status != pb.ConvertFileQueryResponse_COMPLETED.String() &&
			job.Status != pb.ConvertFileQueryResponse_FAILED.String() {}
		done <- true
	}()
	select {
	case <-timeout:
		t.Fatal("timeout waiting for DB status change")
	case <-done:
	}
	assert.Equal(t, pb.ConvertFileQueryResponse_COMPLETED.String(), job.Status, "test should have successfully executed")
	assert.Equal(t, fmt.Sprintf("http://%s.%s/%s/%s", testRegion, testS3Endpoint, testBucketName, res.Id),
		job.CurrUrl, "should have a properly formatted URL")
	assert.GreaterOrEqual(t, time.Now().Unix(), job.LastUpdated.Unix(), "last updated should be recent")
}

func TestConverterServer_ConvertFileQuery_Fail(t *testing.T) {
	// timeout strategy from https://stackoverflow.com/questions/24929790/how-to-set-the-go-timeout-flag-on-go-test
	timeout := time.After(3 * time.Second)
	done := make(chan bool)
	config := testingConfiguration()
	server := converterservice.NewWithConfiguration(toServerConfiguration(config))
	config.S3service.Success = false
	res, err := server.ConvertFile(context.TODO(), testGrpcRequest)
	assert.Nil(t, err, "should not have errored")
	assert.NotNil(t, res, "response should not be nil")
	job, err := config.Db.GetConversion(res.Id)
	go func() {
		for job.Status != pb.ConvertFileQueryResponse_COMPLETED.String() &&
			job.Status != pb.ConvertFileQueryResponse_FAILED.String() {}
		done <- true
	}()
	select {
	case <-timeout:
		t.Fatal("timeout on test case")
	case <-done:
	}
	assert.Equal(t, pb.ConvertFileQueryResponse_FAILED.String(), job.Status, "test should have failed to execute")
	assert.Equal(t, "NONE", job.CurrUrl, "URL should be none")
	assert.GreaterOrEqual(t, time.Now().Unix(), job.LastUpdated.Unix(), "last updated should be recent")
}

func TestConverterServer_ConvertStream(t *testing.T) {
	t.Skip("not implemented")
}