FROM alpine:3.5
MAINTAINER Jimi Olaniyan

WORKDIR /usr/src/app
COPY ./blog ./

RUN ls -lha

EXPOSE 8090
ENTRYPOINT ["./blog"]