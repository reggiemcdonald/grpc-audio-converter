package converterservice

import (
	"errors"
	encodings "github.com/reggiemcdonald/grpc-audio-converter/converterservice/enums"
	"github.com/reggiemcdonald/grpc-audio-converter/pb"
)

type FileConversionRequest struct {
	SourceUrl        string
	SourceEncoding   encodings.Encoding
	DestEncoding     encodings.Encoding
	Id               string
	IncludeExtension bool
}

func NewFileConversionRequest(req *pb.ConvertFileRequest, id string) (*FileConversionRequest, error) {
	if req.SourceUrl == "" {
		return nil, errors.New("request missing required parameter SourceUrl")
	}
	if req.SourceEncoding == req.DestEncoding {
		return nil, errors.New("source and destination encoding are the same")
	}
	sourceEncoding, err := encodings.ToEncoding(int(req.SourceEncoding))
	if err != nil {
		return nil, err
	}
	destEncoding, err := encodings.ToEncoding(int(req.DestEncoding))
	if err != nil {
		return nil, err
	}
	return &FileConversionRequest{
		SourceUrl: req.SourceUrl,
		SourceEncoding: sourceEncoding,
		DestEncoding: destEncoding,
		Id: id,
		// TODO: Add this as a param to the protobuf
		IncludeExtension: false,
	}, nil
}