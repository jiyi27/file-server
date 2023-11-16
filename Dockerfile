FROM golang:alpine

WORKDIR /app
COPY ./ ./
RUN go mod download
RUN go build -o server

#CMD ["./server", "-p", "80", "-tls-port", "443", "-ssl-cert",
#"./conf/cert.pem", "-ssl-key", "./conf/cert.key", "-max", "300"]

# docker build [--platform linux/amd64] -t shwezhu/file-server:v1.0 .
# docker push shwezhu/file-server:v1.0
# docker pull shwezhu/file-server:v1.0
# docker run --name file-server --rm -d -p 80:80 shwezhu/file-server:v1.0 -p 80 -max 120
# sudo docker run -d -p 80:80 -p 443:443 --name file-server --mount source=aaa,target=/app --rm shwezhu/file-server:v1.0
# -p 80 -tls-port 443 -ssl-cert ./conf/cert.pem -ssl-key ./conf/cert.key -max 300