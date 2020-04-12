package converterservice

import (
	"github.com/reggiemcdonald/grpc-audio-converter/pb"
	"fmt"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
)

type server struct {}

func NewConverterService(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen to port %d, caused by %v. Is this port occupied?", port, err)
	}
	s := grpc.NewServer()
	pb.RegisterConverterServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failure! %v", err)
	}
}

func (s *server) ConvertS3File(ctx context.Context, req *pb.ConvertS3FileRequest) (*pb.ConvertS3FileResponse, error) {
	// TODO: Stub
	return &pb.ConvertS3FileResponse{Accepted: false, Id: "ID"}, nil
}

func (s *server) ConvertS3FileQuery(ctx context.Context, req *pb.ConvertS3FileQueryRequest) (*pb.ConvertS3FileQueryResponse, error) {
	// TODO: stub
	return &pb.ConvertS3FileQueryResponse{Id: "ID", Status: 3}, nil
}

func (s *server) ConvertStream(ctx context.Context, req *pb.ConvertStreamRequest) (*pb.ConvertStreamResponse, error) {
	// TODO: stub
	return &pb.ConvertStreamResponse{Buff: []byte{}, Encoding: 0}, nil
}