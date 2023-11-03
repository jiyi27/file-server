## file-server

A tiny http file storage system written in Go, support share file and authentication.

In real development environment, file transfer can be an issue, because there are many files need to be transferred from the personal machine to the development machine and vice versa. With file-server deploying as a server, you can manage (download, upload, delete, create folder) the files easily between different computers easily.

![Go HTTP file server pages](doc/server.gif)

## Compile

```shell
$ go build -o server
```

Will generate executable file "server" in current directory.

## Examples

Start server on port 8080, root directory is `root/` under the project directory:

```shell
./server -p 8080
```

Start server on port 8080, root directory is /usr/share/doc:

```shell
./server -p 8080 -r /usr/share/doc
```

Start server on port 80 and 443, 80 serves for plain HTTP, 443 serves for HTTPS:

```shell
$ ./server -plain-port 80 -tls-port 443 -ssl-cert ~/tls/server.crt -ssl-key ~/tls/server.key
```

For generating TLS Certificate, please refer to: https://gist.github.com/denji/12b3a568f092ab951456

Http Basic Auth:

- listen port: 8080
- protocol: plain HTTP
- username: admin
- password: admin123

```shell
$ ./server -p 8080 -user admin:admin123
```

## Usage

```shell
$ ./server -h
Usage of ./server:
-max int
    maximum size for single file uploads in MB (default 32)
-p int
    (alias for -plain-port)
-plain-port int
    plain http port the server listens on
-r string
    (alias for -root) (default "root/")
-root string
    root directory for the file server (default "root/")
-ssl-cert string
    path to SSL server certificate
-ssl-key string
    path to SSL private key
-tls-port int
    tls port the server listens on, will fail if cert or key is not specified
-user string
    --user <username:password>
    specify user for HTTP Basic Auth (default "admin:admin")
```

## Acknowledgments

- [sgreben/http-file-server](https://github.com/sgreben/http-file-server)
- [mjpclab/go-http-file-server](https://github.com/mjpclab/go-http-file-server)
