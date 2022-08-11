FROM python:3.8-buster as builder

RUN apt-get update
RUN apt install -y ffmpeg

#RUN pip install --no-cache-dir tensorflow==2.3.0
RUN pip install --no-cache-dir spleeter==2.3.0

RUN mkdir /spleeter-scratch
RUN mkdir /youtubedl-scratch

RUN wget https://go.dev/dl/go1.18.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.18.linux-amd64.tar.gz
RUN rm go1.18.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin

WORKDIR /chord-paper-be/worker

COPY ./go.mod ./go.sum ./
COPY ./worker/src/ ./worker/src/

RUN go build -o chord-paper-be-workers ./worker/src/worker.go
 
CMD ["./chord-paper-be-workers"]
