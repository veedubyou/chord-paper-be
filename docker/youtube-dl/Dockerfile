FROM alpine:latest

RUN apk add curl 
RUN mkdir /youtube-dl
RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /home/youtube-dl
RUN chmod +x /home/youtube-dl
