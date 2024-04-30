## variables
GO = go

## build
.PHONY: build
build:
	GO111MODULE=on $(GO) build -v -o _output/file-server ./
	

.PHONY: build-linux
build-linux:
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -v -o _output/file-server-linux-amd64 ./
