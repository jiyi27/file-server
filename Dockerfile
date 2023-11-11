FROM golang:alpine

WORKDIR /app
COPY ./ ./
RUN go mod download
RUN go build -o server

CMD ["./server", "-p", "80"]

# docker build [--platform linux/amd64] -t shwezhu/file-server:v1.0 .
# docker push shwezhu/file-server:v1.0
# docker pull davidzhu/file-server:v1.0
# sudo docker run -d -p 80:80 --rm shwezhu/file-server:v1.0