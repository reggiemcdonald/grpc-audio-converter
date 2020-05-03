package main

import (
	"flag"
	"github.com/joho/godotenv"
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice"
	"log"
)

func main() {
	var (
		port = flag.Int("port", 3000, "port to run service on")
	)
	flag.Parse()
	if err := godotenv.Load(); err != nil {
		log.Printf("failed to load environment config %v", err)
	}
	converterservice.StartConverterService(*port)
}
