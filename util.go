package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
)

func randomString(n int) (string, error) {
	const letters = "0123456789"
	res := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random string: %v", err)
		}
		res[i] = letters[num.Int64()]
	}

	return string(res), nil
}

// formatPath ensures path start with os.PathSeparator and end with no os.PathSeparator.
// And replaces all '/' with os.PathSeparator.
func formatPath(path string) string {
	path = strings.ReplaceAll(path, "/", string(os.PathSeparator))

	path = strings.TrimSuffix(path, string(os.PathSeparator))
	if !strings.HasPrefix(path, string(os.PathSeparator)) {
		path = fmt.Sprintf("%v%v", string(os.PathSeparator), path)
	}

	return path
}

// getScheme returns 'http' or 'https'
func getScheme(r *http.Request) (scheme string) {
	scheme = "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return
}

func getContentType(filepath string) (string, error) {
	ext := path.Ext(filepath)
	ctype := mime.TypeByExtension(ext)
	if len(ctype) > 0 {
		return ctype, nil
	}

	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to get content type:%v", err)
	}

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		return "", err
	}

	ctype = http.DetectContentType(buf[:n])
	return ctype, nil
}
