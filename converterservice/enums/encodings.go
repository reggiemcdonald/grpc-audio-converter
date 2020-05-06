// Encodings enumeration for supported audio types
package enums

import (
	"errors"
)

const (
	WAV codec = iota
	MP4
	MP3
	FLAC
)
var encodingsName = []string{
	"WAV",
	"MP4",
	"MP3",
	"FLAC",
}
var encodingsCodec = []codec{
	WAV,
	MP4,
	MP3,
	FLAC,
}

type codec int

type Encoding interface {
	Name() string
}

func (c codec) Name() string {
	return encodingsName[c]
}

func ToEncoding(enumVal int) (codec, error) {
	if enumVal >= len(encodingsCodec) {
		return -1, errors.New("unsupported audio encoding")
	}
	return encodingsCodec[enumVal], nil
}