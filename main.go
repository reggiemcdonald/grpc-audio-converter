// Runs the server using a configuration from environment variables
package main

import (
	"github.com/joho/godotenv"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice"
	"log"
	"os"
	"strconv"
)

/*
 * Returns the required environment variable if it is present
 */
func getEnvAsString(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("missing required environment variable %s", key)
	}
	return value
}

func getEnvAsInt(key string) int {
	stringValue := getEnvAsString(key)
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
	port       := getEnvAsInt("PORT")
	bucketName := getEnvAsString("BUCKET_NAME")
	s3endpoint := getEnvAsString("S3_ENDPOINT")
	region     := getEnvAsString("REGION")
	isDev      := getEnvAsString("DEV") == "true"
	dbUser     := getEnvAsString("POSTGRES_USER")
	dbPass     := getEnvAsString("POSTGRES_PASSWORD")
	db         := converterservice.NewFileConverterData(dbUser, dbPass)
	return &converterservice.ConverterServerConfig{
		Port: port,
		BucketName: bucketName,
		S3endpoint: s3endpoint,
		S3Region: region,
		Db: db,
		IsDev: isDev,
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("failed to load environment config %v", err)
	}
	server := converterservice.NewWithConfiguration(defaultConfiguration())
	converterservice.Start(server)
}
