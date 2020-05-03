package converterservice

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/reggiemcdonald/grpc-audio-converter/pb"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

type ConverterServer struct {
	fileConverter FileConverterService
	db            FileConverterDataService
}

func newDefaultFileConverterConfiguration(db FileConverterDataService) FileConverterConfiguration {
	bucketName := os.Getenv("BUCKET_NAME")
	s3endpoint := os.Getenv("S3_ENDPOINT")
	region     := os.Getenv("REGION")
	isDev      := os.Getenv("DEV") == "true"
	return FileConverterConfiguration{
		bucketName: bucketName,
		s3endpoint: s3endpoint,
		region:     region,
		db:         db,
		isDev:      isDev,
	}
}

/*
 * Creates a new converter service instance
 */
func NewConverterServer() *ConverterServer {
	dbUser     := os.Getenv("POSTGRES_USER")
	dbPass     := os.Getenv("POSTGRES_PASSWORD")
	db         := NewFileConverterData(dbUser, dbPass)
	config     := newDefaultFileConverterConfiguration(db)
	return &ConverterServer{
		fileConverter: NewFileConverter(config),
		db: db,
	}
}

func StartConverterService(port int) {
	log.Println("Starting service...")
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen to port %d, caused by %v. Is this port occupied?", port, err)
	}
	s := grpc.NewServer()
	converterServer := NewConverterServer()
	pb.RegisterConverterServiceServer(s, converterServer)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failure! %v", err)
	}
}

func (s *ConverterServer) ConvertFile(ctx context.Context, req *pb.ConvertFileRequest) (*pb.ConvertFileResponse, error) {
	uuid := uuid.New().String()
	go s.fileConverter.ConvertFile(req, uuid)
	return &pb.ConvertFileResponse{Accepted: true, Id: uuid}, nil
}

func (s *ConverterServer) ConvertFileQuery(ctx context.Context, req *pb.ConvertFileQueryRequest) (*pb.ConvertFileQueryResponse, error) {
	db := s.db
	job, err := db.GetConversion(req.Id)
	if err != nil {
		log.Printf("failed to get %s, encountered %v", req.Id, err)
		return nil, errors.New(fmt.Sprintf("failed to get %s", req.Id))
	}
	return &pb.ConvertFileQueryResponse{
		Id: job.id,
		Status: pb.ConvertFileQueryResponse_Status(pb.ConvertFileQueryResponse_Status_value[job.status]),
		Url: job.currUrl,
	}, nil
}

func (s *ConverterServer) ConvertStream(ctx context.Context, req *pb.ConvertStreamRequest) (*pb.ConvertStreamResponse, error) {
	// TODO: stub
	return &pb.ConvertStreamResponse{Buff: []byte{}, Encoding: 0}, nil
}