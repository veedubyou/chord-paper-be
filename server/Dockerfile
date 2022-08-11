FROM golang:1.18.3-buster as builder

WORKDIR /chord-paper-be/server

COPY ./go.mod ./go.sum ./
COPY ./server/src/ ./server/src/

RUN go build -o chord-paper-be ./server/src/server.go

CMD ["./chord-paper-be"]

