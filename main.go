package main

import (
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"strings"
)

func main() {
	u, err := url.Parse("https://example.com/foobar/a.txt?download")
	if err != nil {
		log.Fatal(err)
	}
	reqPath := strings.TrimSuffix(u.Path, "/")
	reqPath = strings.ReplaceAll(reqPath, "/", string(filepath.Separator))
	osPath := filepath.Join("files", reqPath)
	fmt.Println(osPath)
}
