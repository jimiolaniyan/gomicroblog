FROM golang:1.12.12 AS builder
MAINTAINER Jimi Olaniyan

ENV GO111MODULE=on
WORKDIR /go/src/github.com/jimiolaniyan/gomicroblog/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o app api/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /go/src/github.com/jimiolaniyan/gomicroblog/app .

EXPOSE 8090
ENTRYPOINT ["./app"]