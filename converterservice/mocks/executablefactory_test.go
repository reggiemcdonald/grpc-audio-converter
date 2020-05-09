package mocks

import (
	"fmt"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice/enums"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice/fileconverter"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewMockExecutableFactory(t *testing.T) {
	factory := NewMockExecutableFactory()
	assert.NotNil(t, factory)
}

func TestMockExecutableFactory_Build(t *testing.T) {
	sourceUrl := "test-url"
	sourceEncoding := enums.MP3
	destEncoding := enums.MP4
	id := "test-id"
	includeExtension := false
	req := &fileconverter.FileConversionRequest{
		SourceUrl: sourceUrl,
		SourceEncoding: sourceEncoding,
		DestEncoding: destEncoding,
		Id: id,
		IncludeExtension: includeExtension,
	}
	t.Run("success=true", func(t *testing.T) {
		factory := NewMockExecutableFactory()
		job := &fileconverter.ConversionAttributes{
			Request: req,
		}
		cmd := factory.Build(job)
		assert.Same(t, req, job.Request)
		if err := cmd.Start(); err != nil {
			t.Error(err.Error())
		}
		assert.Equal(t, fmt.Sprintf("mock executable for %s", id), cmd.String())
	})
	t.Run("success=false", func(t *testing.T) {
		factory := NewMockExecutableFactory()
		job := &fileconverter.ConversionAttributes{
			Request: req,
		}
		factory.Success = false
		cmd := factory.Build(job)
		assert.Same(t, req, job.Request)
		if err := cmd.Start(); err == nil {
			t.Error("expected error but got none")
		}
		assert.Equal(t, fmt.Sprintf("mock executable for %s", id), cmd.String())
	})
}
