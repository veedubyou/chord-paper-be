FROM golang:1.18.3-buster as builder

WORKDIR /chord-paper-be/src/server

COPY ./go.mod ./go.sum ./
COPY ./src/ ./src/

RUN go build -o chord-paper-be ./src/server/server.go

CMD ["./chord-paper-be"]

