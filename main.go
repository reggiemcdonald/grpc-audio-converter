package main

import (
	"github.com/reggiemcdonald/grpc-audio-converter/converterservice"
)

const (
	port = 3000
)


func main() {
	converterservice.NewConverterService(port)
}