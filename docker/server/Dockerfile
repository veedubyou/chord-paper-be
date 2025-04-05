FROM golang:1.18.3-buster as builder

RUN apt-get update
RUN apt install -y python3-pip

RUN pip install --no-cache-dir syncedlyrics==1.0.1

WORKDIR /chord-paper-be/src/server

COPY ./go.mod ./go.sum ./
COPY ./src/ ./src/

RUN go build -o chord-paper-be ./src/server/server.go

CMD ["./chord-paper-be"]

