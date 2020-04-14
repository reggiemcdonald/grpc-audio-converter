package main

import (
	"github.com/joho/godotenv"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice"
	"log"
	"os"
	"strconv"
)


func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("failed to load environment config %v", err)
	}
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Println("Missing environment parameter PORT, defaulting to :3000 ...")
		port = 3000
	}
	converterservice.NewConverterService(port)
}