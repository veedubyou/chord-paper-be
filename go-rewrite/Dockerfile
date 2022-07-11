FROM golang:1.18.3-buster as builder

WORKDIR /chord-paper-be/go-rewrite

COPY go.mod go.sum ./
COPY src/ ./src/

RUN go build -o chord-paper-be-go ./src/main.go

CMD ["./chord-paper-be-go"]

