package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleUpload_FileTooLarge(t *testing.T) {
	s := &server{maxFileSize: 1} // 1 MB
	currentDir := "./test_dir"
	_ = os.Mkdir(currentDir, 0750)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "test.txt")
	// Write more than 1MB of data
	_, _ = io.Copy(part, bytes.NewBuffer(make([]byte, 2*1024*1024)))
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	errs := s.handleUpload(w, req, currentDir)

	if len(errs) == 0 || errs[0].Message != "copy file: http: request body too large" {
		t.Errorf("Expected 'http: request body too large' error, got %v", errs[0].Message)
	}

	_, err := os.Stat(filepath.Join(currentDir, "test.txt"))
	if !os.IsNotExist(err) { // new created file should be removed, because it's too large
		t.Errorf("Expected file to not exist")
	}

	_ = os.RemoveAll(currentDir)
}

// ---------------------benchmark---------------------
// go test -run=xxx -bench 'BenchmarkDirectRead|BenchmarkLimitedRead'
var sampleData []byte

func init() {
	dataSize := int64(1024 * 1024 * 5) // 5MB
	sampleData = make([]byte, dataSize)
	for i := range sampleData {
		sampleData[i] = byte(i % 255)
	}
}

func getDataSource() io.Reader {
	return &multipart.Part{}
}

func BenchmarkDirectRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dataSource := getDataSource()
		_, _ = io.Copy(io.Discard, dataSource)
	}
}

func BenchmarkLimitedRead(b *testing.B) {
	limit := int64(1024 * 1024 * 10) // 10MB
	for i := 0; i < b.N; i++ {
		dataSource := getDataSource()
		limitedReader := &LimitedReader{R: dataSource, N: limit}
		_, _ = io.Copy(io.Discard, limitedReader)
	}
}

type LimitedReader struct {
	R io.Reader
	N int64
}

// for test reason, very simple
func (l *LimitedReader) Read(p []byte) (n int, err error) {
	n, err = l.R.Read(p)
	return
}
