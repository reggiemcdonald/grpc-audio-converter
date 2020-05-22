// The ConverterServer
package converterservice

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	db "github.com/reggiemcdonald/grpc-audio-converter/converterservice/db"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice/fileconverter"
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
	fileConverter fileconverter.Converter
	repo          db.FileConverterRepository
	config        *ConverterServerConfig
	queue         FileConverterJobQueue
}

/*
 * The configuration for the converter server
 */
type ConverterServerConfig struct {
	Concurrency       int
	Db                db.FileConverterRepository
	ExecutableFactory fileconverter.ExecutableFactory
	QueueSize         int
	Port              int
	S3service         fileconverter.FileUploader
}

type converterServiceJob struct {
	request *fileconverter.FileConversionRequest
	converter fileconverter.Converter
}

/*
 * Creates a new converter service instance
 */
func NewWithConfiguration(config *ConverterServerConfig) *ConverterServer {
	queue := NewJobQueue(&JobQueueConfiguration{
		Concurrency: config.Concurrency,
		QueueSize: config.QueueSize,
	})
	if err := queue.Start(); err != nil {
		log.Fatalf("could not start job queue, encountered %v", err)
	}
	return &ConverterServer{
		fileConverter: fileconverter.New(&fileconverter.ConverterImplementation{
			S3service: config.S3service,
			Db: config.Db,
			ExecutableFactory: config.ExecutableFactory,
		}),
		repo:   config.Db,
		config: config,
		queue: queue,
	}
}

func (s *ConverterServer) newJob(request *fileconverter.FileConversionRequest) FileConverterJob {
	return &converterServiceJob{
		converter: s.fileConverter,
		request: request,
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
	request, err := fileconverter.NewFileConversionRequest(req, id)
	if err != nil {
		if _, dbErr := s.repo.FailConversion(id); dbErr != nil {
			log.Printf("failed to update DB with failure, encountered %v", dbErr)
		}
		return nil, err
	}
	if _, err := s.repo.NewRequest(id); err != nil {
		return nil, errors.New("an internal error occurred")
	}
	if err = s.queue.Enqueue(s.newJob(request)); err != nil {
		log.Printf("failed to add job to queue, encountered %v", err)
		return nil, errors.New("an internal error occurred")
	}
	return &pb.ConvertFileResponse{Accepted: true, Id: id}, nil
}

func (s *ConverterServer) ConvertFileQuery(ctx context.Context, req *pb.ConvertFileQueryRequest) (*pb.ConvertFileQueryResponse, error) {
	job, err := s.repo.GetConversion(req.Id)
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

func (j *converterServiceJob) Start() {
	j.converter.ConvertFile(j.request)
}