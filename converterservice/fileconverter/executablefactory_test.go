package fileconverter

import (
	"errors"
	"fmt"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice/enums"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func trimCommand(command string) (string, error) {
	start := strings.Index(command, ffmpeg)
	if start < 0 {
		return "", errors.New("not found")
	}
	return command[start:], nil
}

func TestDefaultExecutableFactory_Build(t *testing.T) {
	factory := newDefaultExecutableFactory()
	sourceUrl := "test-url"
	sourceEncoding := enums.MP3
	id := "test-id"
	includeExtension := true
	// test default
	t.Run("encoding=WAV", func(t *testing.T) {
		destEncoding := enums.WAV
		job := &ConversionAttributes{
			Request: &FileConversionRequest{
				SourceUrl: sourceUrl,
				SourceEncoding: sourceEncoding,
				DestEncoding: destEncoding,
				Id: id,
				IncludeExtension: includeExtension,
			},
		}
		cmd := factory.Build(job)
		// Should not have changed request
		assert.Equal(t, sourceUrl, job.Request.SourceUrl)
		assert.Equal(t, sourceEncoding, job.Request.SourceEncoding)
		assert.Equal(t, destEncoding, job.Request.DestEncoding)
		assert.Equal(t, id, job.Request.Id)
		assert.True(t, job.Request.IncludeExtension)
		// Check the command
		assert.Equal(t,
			fmt.Sprintf("/tmp/%s.%s", id, strings.ToLower(job.Request.DestEncoding.Name())),
			job.TmpFile)
		command, err := trimCommand(cmd.String())
		if err != nil {
			t.Error("command does not match")
		}
		assert.Equal(t,
			fmt.Sprintf(
				"ffmpeg -f %s -i %s -map 0:0 -f %s %s",
				sourceEncoding.Name(),
				sourceUrl,
				destEncoding.Name(),
				fmt.Sprintf("/tmp/%s.%s", id, strings.ToLower(destEncoding.Name()))),
				command)
	})
	// test MP4
	t.Run("encoding=MP4", func(t *testing.T) {
		destEncoding := enums.MP4
		job := &ConversionAttributes{
			Request: &FileConversionRequest{
				SourceUrl: sourceUrl,
				SourceEncoding: sourceEncoding,
				DestEncoding: destEncoding,
				Id: id,
				IncludeExtension: includeExtension,
			},
		}
		cmd := factory.Build(job)
		// Should not have changed request
		assert.Equal(t, sourceUrl, job.Request.SourceUrl)
		assert.Equal(t, sourceEncoding, job.Request.SourceEncoding)
		assert.Equal(t, destEncoding, job.Request.DestEncoding)
		assert.Equal(t, id, job.Request.Id)
		assert.True(t, job.Request.IncludeExtension)
		// Check command
		commandString, err := trimCommand(cmd.String())
		if err != nil {
			t.Error("command does not match")
		}
		assert.Equal(t,
			fmt.Sprintf(
				"ffmpeg -f %s -i %s -map 0:0 -f %s %s",
				sourceEncoding.Name(),
				sourceUrl,
				destEncoding.Name(),
				fmt.Sprintf("/tmp/%s.m4a", id)),
			commandString)
	})
}
