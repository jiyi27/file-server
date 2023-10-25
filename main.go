package main

import (
	"log"
	"net/http"
)

func main() {
	srv := server{
		root:           "root",
		rootAssetsPath: "template",
		maxFileSize:    1 << 25,
	}
	if err := http.ListenAndServe(":8080", &srv); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
