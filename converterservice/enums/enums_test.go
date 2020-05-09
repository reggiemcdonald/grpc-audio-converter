package enums

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStatus_Name(t *testing.T) {
	assert.Equal(t, "QUEUED", QUEUED.Name())
	assert.Equal(t, "CONVERTING", CONVERTING.Name())
	assert.Equal(t, "COMPLETED", COMPLETED.Name())
	assert.Equal(t, "FAILED", FAILED.Name())
}

func TestStatus_Value(t *testing.T) {
	assert.Equal(t, 0, QUEUED.Value())
	assert.Equal(t, 1, CONVERTING.Value())
	assert.Equal(t, 2, COMPLETED.Value())
	assert.Equal(t, 3, FAILED.Value())
}

func TestStatusFromEnumValue(t *testing.T) {
	for _, status := range statuses {
		s, err := StatusFromEnumValue(status.Value())
		assert.Equal(t, status, s)
		assert.Nil(t, err)
	}
	_, err := StatusFromEnumValue(6)
	assert.NotNil(t, err)
}

func TestEncoding_Name(t *testing.T) {
	assert.Equal(t, "WAV", WAV.Name())
	assert.Equal(t, "MP4", MP4.Name())
	assert.Equal(t, "MP3", MP3.Name())
	assert.Equal(t, "FLAC", FLAC.Name())
}

func TestEncoding_Value(t *testing.T) {
	assert.Equal(t, 0, WAV.Value())
	assert.Equal(t, 1, MP4.Value())
	assert.Equal(t, 2, MP3.Value())
	assert.Equal(t, 3, FLAC.Value())
}

func TestFromEnumToEncoding(t *testing.T) {
	for _, encoding := range encodings {
		s, err := EncodingFromEnumValue(encoding.Value())
		assert.Equal(t, encoding, s)
		assert.Nil(t, err)
	}
	_, err := EncodingFromEnumValue(6)
	assert.NotNil(t, err)
}


