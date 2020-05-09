// Encodings enumeration for supported audio types
package enums

import (
	"errors"
)

const (
	WAV encoding = iota
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
var encodings = []encoding{
	WAV,
	MP4,
	MP3,
	FLAC,
}

type encoding int

type Encoding interface {
	Name()  string
	Value() int
}

func (c encoding) Name() string {
	return encodingsName[c]
}

func (c encoding) Value() int {
	return int(c)
}

func EncodingFromEnumValue(enumVal int) (encoding, error) {
	if enumVal >= len(encodings) {
		return -1, errors.New("unsupported audio encoding")
	}
	return encodings[enumVal], nil
}