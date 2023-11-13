FROM golang:alpine

WORKDIR /app
COPY ./ ./
RUN go mod download
RUN go build -o server

ENTRYPOINT ["./server"]

# docker build [--platform linux/amd64] -t shwezhu/file-server:v1.0 .
# docker push shwezhu/file-server:v1.0
# docker pull shwezhu/file-server:v1.0
# docker run --name file-server --rm -d -p 80:80 shwezhu/file-server:v1.0 -p 80 -max 120