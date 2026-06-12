FROM pw1124/chord-be-workers:preinstall-latest

RUN mkdir /spleeter-scratch
RUN mkdir /demucs-scratch
RUN mkdir /youtubedl-scratch

WORKDIR /chord-paper-be/src/worker

COPY ./go.mod ./go.sum ./
COPY ./src/ ./src/

RUN go build -o chord-paper-be-workers ./src/worker/worker.go
 
CMD ["./chord-paper-be-workers"]
