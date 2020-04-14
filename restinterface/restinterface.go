package main

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/reggiemcdonald/grpc-audio-converter/pb"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"strings"
)

func stringToEncoding(encoding string) (int, error) {
	switch enc := strings.ToUpper(encoding); enc {
	case "WAV":
		return 0,nil
	case "M4A":
		return 1,nil
	case "MP3":
		return 2,nil
	case "FLAC":
		return 3,nil
	default:
		return -1,errors.New("invalid encoding specified")
	}
}

type body struct {
	SourceUrl string `json:"sourceUrl"`
}

func main() {
	conn, err := grpc.Dial("localhost:3000", grpc.WithInsecure())
	if err != nil {
		log.Fatal("Failed to connect to the grpc service")
	}
	client := pb.NewConverterServiceClient(conn)
	r := gin.Default()
	r.GET("/convert-file", func(c *gin.Context) {
		log.Println("Received request")
		// TODO
	})
	r.POST("/convert-file", func(c *gin.Context) {
		var b body
		if err := c.ShouldBindJSON(&b); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing or improperly formatted request body"})
			return
		}
		srcEncoding, errSrc := stringToEncoding(c.Query("src"))
		destEncoding, errDst := stringToEncoding(c.Query("dest"))
		if errSrc != nil || errDst != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid params"})
			return
		}
		res, err := client.ConvertFile(c, &pb.ConvertFileRequest{
			SourceUrl: b.SourceUrl,
			SourceEncoding: pb.Encoding(srcEncoding),
			DestEncoding: pb.Encoding(destEncoding),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"id": res.Id})
	})
	if err = r.Run(":4000"); err != nil {
		log.Fatal("Failed to start")
	}
}