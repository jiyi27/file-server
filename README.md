# file-server

A tiny http file storage system written in Go, support share file and authentication.

In real development environment, file transfer can be an issue, because there are many files need to be transferred from the personal machine to the development machine and vice versa. With file-server deploying as a server, you can manage (download, upload, delete, create folder) the files easily between different computers easily.

![Go HTTP file server pages](doc/server.gif)

## Features

- Friendly UI
- Zero dependencies
- Single executable file
- Upload multiple files one time
- Stream file to server, friendly for uploading large file
- Create folder at the current directory
- Can generate shared download link for a file
- Can specify authentication for custom directories

## Compile

```shell
$ go build -o file_server ./
```

## Examples

Start server on port 8080, all the files will be saved in the `./root` directory by default:

```shell
nohup ./server -http 8080 &
```

> `nohup` detaches the process from the terminal and redirects its output to nohup.out

Http Basic Auth:

- Listen port: 8080
- Require authentication for accessing all files

```shell
./server -http 8080 -auth :admin:adminpw
# or
./server -http 8080 -auth /:admin:adminpw
```

Another complicated example:

- Listen port: 80 & 443
- Requires authentication for url `/abc`
  - username: `user1`, password: `user1pw`

- Requires authentication for url `/aaa/bbb`
  - username: `user2`, password: `user2pw`

```shell
$./server -auth /abc:user1:user1pw -auth /aaa/bbb:user2:user2pw -http 80 -https 443 -ssl-cert ./conf/cert.pem -ssl-key ./conf/cert.key
```

Start server on port 80 and 443, with SSL enabled, and authentication required for all files, username: `1132`, password: `1132`, and save the log to `nohup.out`, ssl certificate and key are in the `../tls` directory:

```shell
nohup ./file_server -auth :1132:1132 -http 80 -https 443 -ssl-key ../tls/cert.key -ssl-cert ../tls/cert.pem &
```

## Usage

```shell
❯ ./server -h
Usage of ./server:
    -auth value
        -auth <path:username:password>
        example: /:admin:adminpw
    -http int
        Enable HTTP server on the specified port, default is 80
    -https int
        Enable HTTPS server on the specified port, default is 443
    -max int
        Specify maximum size for single file uploads in MB (default 32)
    -r string
        (alias for -root) (default "./root")
    -root string
        Specify the root directory to save files (default "./root")
    -ssl-cert string
        Specify the path of SSL certificate
    -ssl-key string
        Specify the path of SSL private key
```

## Acknowledgments

- [sgreben/http-file-server](https://github.com/sgreben/http-file-server)
- [mjpclab/go-http-file-server](https://github.com/mjpclab/go-http-file-server)
