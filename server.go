package main

type server struct {
	root        string
	maxFileSize int64 // = 1 << 25 // 32MB
}
