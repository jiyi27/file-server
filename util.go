package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//func randomString(n int) (string, error) {
//	const letters = "0123456789"
//	res := make([]byte, n)
//	for i := 0; i < n; i++ {
//		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
//		if err != nil {
//			return "", fmt.Errorf("failed to generate random string: %v", err)
//		}
//		res[i] = letters[num.Int64()]
//	}
//
//	return string(res), nil
//}

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

// pathDepth calculates the depth of a directory in the file path.
func pathDepth(path string) int {
	if path == "/" {
		return 0
	}

	return len(strings.Split(filepath.Clean(path), string(filepath.Separator))) - 1
}

//// getScheme returns 'http' or 'https'
//func getScheme(r *http.Request) (scheme string) {
//	scheme = "http"
//	if r.TLS != nil {
//		scheme = "https"
//	}
//	return
//}

//func getContentType(filepath string) (string, error) {
//	ext := path.Ext(filepath)
//	ctype := mime.TypeByExtension(ext)
//	if len(ctype) > 0 {
//		return ctype, nil
//	}
//
//	file, err := os.Open(filepath)
//	if err != nil {
//		return "", fmt.Errorf("failed to get content type:%v", err)
//	}
//
//	buf := make([]byte, 512)
//	n, err := file.Read(buf)
//	if err != nil {
//		return "", err
//	}
//
//	ctype = http.DetectContentType(buf[:n])
//	return ctype, nil
//}

func getAvailableName(fileDir, fileName string) string {
	filePath := path.Join(fileDir, fileName)
	// file already exists, generate a new name.
	if _, err := os.Stat(filePath); err == nil {
		fileName =
			strings.Split(fileName, ".")[0] +
				"_" +
				strconv.FormatInt(time.Now().UnixNano(), 10) +
				filepath.Ext(fileName)
	}
	// no such file, use the original name.
	return fileName
}

func generateHash(input string) string {
	hashes := sha256.New()
	hashes.Write([]byte(input))
	hash := hashes.Sum(nil)
	return hex.EncodeToString(hash) // 将哈希值转换为十六进制字符串
}

func serveHTTP(port int, s *server) {
	log.Printf("listening http on %v", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), s))
}

func serveHTTPS(port int, cert, key string, s *server) {
	log.Printf("listening https on %v", port)
	log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%v", port), cert, key, s))
}
