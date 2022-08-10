FROM golang:1.18.3-buster as builder

WORKDIR /chord-paper-be/server

COPY ./server/go.mod ./server/go.sum ./
COPY ./server/src/ ./src/

RUN go build -o chord-paper-be-go ./src/server.go

CMD ["./chord-paper-be-go"]

