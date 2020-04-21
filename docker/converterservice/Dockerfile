# Multi-stage Build for Converter Service
FROM znly/protoc:latest
RUN mkdir /pb
COPY ./pb /pb
WORKDIR /pb
RUN protoc -I . --go_out=plugins=grpc:. ./*.proto

FROM golang:latest
RUN mkdir /app
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux go build -o converterservice-app .

FROM jrottenberg/ffmpeg:4.1-alpine
WORKDIR /app
COPY --from=1 /app .
ENTRYPOINT ["./converterservice-app"]