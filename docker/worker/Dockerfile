FROM python:3.9-buster as builder

RUN apt-get update
RUN apt install -y ffmpeg

RUN pip install --no-cache-dir spleeter==2.3.0
RUN pip install --no-cache-dir demucs==4.0.1
RUN pip install --no-cache-dir syncedlyrics==1.0.1

RUN mkdir /spleeter-scratch
RUN mkdir /demucs-scratch
RUN mkdir /youtubedl-scratch

RUN wget https://go.dev/dl/go1.18.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.18.linux-amd64.tar.gz
RUN rm go1.18.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin:/shared

WORKDIR /chord-paper-be/src/worker

COPY ./go.mod ./go.sum ./
COPY ./src/ ./src/

RUN go build -o chord-paper-be-workers ./src/worker/worker.go
 
CMD ["./chord-paper-be-workers"]
