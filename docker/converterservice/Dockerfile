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
COPY --from=0 /pb/* pb/
RUN CGO_ENABLED=0 GOOS=linux go build -o converterservice-app .

FROM jrottenberg/ffmpeg:4.1-alpine
WORKDIR /app
COPY --from=1 /app .
ENTRYPOINT ["./converterservice-app"]