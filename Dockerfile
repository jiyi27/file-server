FROM golang:alpine

WORKDIR /app
COPY ./ ./
RUN go mod download
RUN go build -o server

ENTRYPOINT ["./server"]

# docker build [--platform linux/amd64] -t shwezhu/file-server:v1.0 .
# docker push shwezhu/file-server:v1.0
# docker pull davidzhu/file-server:v1.0
# sudo docker run -p 9000:9000 --rm file-server:v1.0 -p 9000 -max-size 2000