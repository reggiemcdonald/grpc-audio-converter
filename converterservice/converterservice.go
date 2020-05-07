// The ConverterServer
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
)

/*
 * The parts of the converter server
 */
type ConverterServer struct {
	fileConverter FileConverterService
	db            FileConverterDataRepository
	config        *ConverterServerConfig
}

/*
 * The configuration for the converter server
 */
type ConverterServerConfig struct {
	Db                FileConverterDataRepository
	ExecutableFactory ExecutableFactory
	Port              int
	S3service         S3Service
}

/*
 * Creates a new converter service instance
 */
func NewWithConfiguration(config *ConverterServerConfig) *ConverterServer {
	return &ConverterServer{
		fileConverter: NewFileConverter(&FileConverterConfiguration{
			S3service: config.S3service,
			Db:         config.Db,
			ExecutableFactory: config.ExecutableFactory,
		}),
		db: config.Db,
		config: config,
	}
}

func Start(server *ConverterServer) {
	log.Println("Starting service...")
	port := server.config.Port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Failed to listen to port %d, caused by %v. Is this port occupied?", port, err)
	}
	s := grpc.NewServer()
	pb.RegisterConverterServiceServer(s, server)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failure! %v", err)
	}
}

func (s *ConverterServer) ConvertFile(ctx context.Context, req *pb.ConvertFileRequest) (*pb.ConvertFileResponse, error) {
	id := uuid.New().String()
	if _, err := s.db.NewRequest(id); err != nil {
		return &pb.ConvertFileResponse{Accepted: false, Id: id}, errors.New("an internal error occurred")
	}
	request, err := NewFileConversionRequest(req, id)
	if err != nil {
		if _, dbErr := s.db.FailConversion(id); dbErr != nil {
			log.Printf("failed to update DB with failure, encountered %v", dbErr)
		}
		return &pb.ConvertFileResponse{Accepted: false, Id: id}, err
	}
	go s.fileConverter.ConvertFile(request)
	return &pb.ConvertFileResponse{Accepted: true, Id: id}, nil
}

func (s *ConverterServer) ConvertFileQuery(ctx context.Context, req *pb.ConvertFileQueryRequest) (*pb.ConvertFileQueryResponse, error) {
	db := s.db
	job, err := db.GetConversion(req.Id)
	if err != nil {
		log.Printf("failed to get %s, encountered %v", req.Id, err)
		return nil, errors.New(fmt.Sprintf("failed to get %s", req.Id))
	}
	return &pb.ConvertFileQueryResponse{
		Id: job.Id,
		Status: pb.ConvertFileQueryResponse_Status(pb.ConvertFileQueryResponse_Status_value[job.Status]),
		Url: job.CurrUrl,
	}, nil
}

func (s *ConverterServer) ConvertStream(ctx context.Context, req *pb.ConvertStreamRequest) (*pb.ConvertStreamResponse, error) {
	// TODO: stub
	return &pb.ConvertStreamResponse{Buff: []byte{}, Encoding: 0}, nil
}