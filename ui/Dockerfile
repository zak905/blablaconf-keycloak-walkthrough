FROM golang:1.16-alpine

RUN apk add --no-cache git 

RUN GO111MODULE=off go get github.com/oxequa/realize

WORKDIR /ui


CMD ["realize", "start", "-n", "ui"]
