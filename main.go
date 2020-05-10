// Runs the server using a configuration from environment variables
package main

import (
	"github.com/joho/godotenv"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice"
	db "github.com/reggiemcdonald/grpc-audio-converter/converterservice/db"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice/fileconverter"
	"log"
	"os"
	"strconv"
)

/*
 * Returns the required environment variable if it is present
 */
func getRequiredEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("missing required environment variable %s", key)
	}
	return value
}

func getEnvWithDefault(key string, def string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return def
	}
	return value
}

func getRequiredEnvAsInt(key string) int {
	stringValue := getRequiredEnv(key)
	numericValue, err := strconv.Atoi(stringValue)
	if err != nil {
		log.Fatalf("invalid env variable %s, encountered %v", key, err)
	}
	return numericValue
}

/*
 * Returns the server configuration from environment variables
 */
func defaultConfiguration() *converterservice.ConverterServerConfig{
	port       := getRequiredEnvAsInt("PORT")
	bucketName := getRequiredEnv("BUCKET_NAME")
	region     := getRequiredEnv("REGION")
	s3endpoint := getEnvWithDefault("S3_ENDPOINT", "")
	isDev      := getEnvWithDefault("DEV", "false") == "true"
	dbUser     := getRequiredEnv("POSTGRES_USER")
	dbPass     := getRequiredEnv("POSTGRES_PASSWORD")
	repo       := db.NewFromCredentials(dbUser, dbPass)
	var s3Service fileconverter.FileUploader
	if isDev {
		s3Service = fileconverter.NewLocalFileUploader(region, s3endpoint, bucketName)
	} else {
		s3Service = fileconverter.NewS3FileUploader(region, s3endpoint, bucketName)
	}
	return &converterservice.ConverterServerConfig{
		Port:      port,
		Db:        repo,
		S3service: s3Service,
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("failed to load environment config %v", err)
	}
	server := converterservice.NewWithConfiguration(defaultConfiguration())
	converterservice.Start(server)
}
