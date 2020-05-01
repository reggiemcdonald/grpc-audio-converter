# Multi-stage build for the rest interface
FROM znly/protoc:latest
RUN mkdir /pb
COPY ./pb /pb
WORKDIR /pb
RUN protoc -I . --go_out=plugins=grpc:. ./*.proto

FROM golang:latest
RUN mkdir /app
WORKDIR /app
COPY . /app
COPY --from=0 /pb/*.pb.go ./pb
RUN CGO_ENABLED=0 GOOS=linux go build -o rest-app restinterface/restinterface.go
ENTRYPOINT ["./rest-app"]