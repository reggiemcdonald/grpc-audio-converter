package fileconverter

import (
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice/enums"
	"github.com/reggiemcdonald/grpc-audio-converter/pb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewFileConversionRequest(t *testing.T) {
	id := "test-id"
	sourceUrl := "test-url"
	req := &pb.ConvertFileRequest{
		SourceUrl: sourceUrl,
		SourceEncoding: pb.Encoding_WAV,
		DestEncoding: pb.Encoding_MP3,
	}
	internalRequest, err := NewFileConversionRequest(req, id)
	assert.Nil(t, err)
	assert.NotNil(t, internalRequest)
	assert.Equal(t, id, internalRequest.Id)
	assert.Equal(t, sourceUrl, internalRequest.SourceUrl)
	assert.Equal(t, enums.WAV, internalRequest.SourceEncoding)
	assert.Equal(t, enums.MP3, internalRequest.DestEncoding)
	// TODO: Parameterize this
	assert.False(t, internalRequest.IncludeExtension)
}
