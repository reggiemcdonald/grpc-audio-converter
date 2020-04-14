package converterservice

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/reggiemcdonald/grpc-audio-converter/pb"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

type server struct {
	fileConverter *FileConverter
}

func NewConverterService(port int) {
	log.Println("Starting service...")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen to port %d, caused by %v. Is this port occupied?", port, err)
	}
	s := grpc.NewServer()
	pb.RegisterConverterServiceServer(s, &server{
		fileConverter: NewFileConverter(os.Getenv("S3_ENDPOINT")),
	})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failure! %v", err)
	}
}

func (s *server) ConvertFile(ctx context.Context, req *pb.ConvertFileRequest) (*pb.ConvertFileResponse, error) {
	// TODO: Create the ID for the request
	uuid := uuid.New().String()
	go s.fileConverter.ConvertFile(req, uuid)
	return &pb.ConvertFileResponse{Accepted: true, Id: uuid}, nil
}

func (s *server) ConvertFileQuery(ctx context.Context, req *pb.ConvertFileQueryRequest) (*pb.ConvertFileQueryResponse, error) {
	// TODO: stub
	log.Println("HERE")
	return &pb.ConvertFileQueryResponse{Id: "ID", Status: 3}, nil
}

func (s *server) ConvertStream(ctx context.Context, req *pb.ConvertStreamRequest) (*pb.ConvertStreamResponse, error) {
	// TODO: stub
	return &pb.ConvertStreamResponse{Buff: []byte{}, Encoding: 0}, nil
}