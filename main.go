package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// AUTH_USERNAME=admin AUTH_PASSWORD=778899 PORT=8080 go run .
func main() {
	srv := server{
		domain:         "http://localhost:8080",
		root:           "root",
		rootAssetsPath: "template",
		maxFileSize:    1 << 25, // 32MB
		filesPathToId:  make(map[string]string),
		filesIdToPath:  make(map[string]string),
	}
	srv.auth.username = os.Getenv("AUTH_USERNAME")
	srv.auth.password = os.Getenv("AUTH_PASSWORD")
	port := os.Getenv("PORT")

	if srv.auth.username == "" {
		log.Fatal("basic auth username must be provided")
	}
	if srv.auth.password == "" {
		log.Fatal("basic auth password must be provided")
	}

	log.Printf("starting server on %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), &srv))
}
