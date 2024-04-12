package main

import (
	"net/http"
	_ "net/http/pprof"
)

func main() {
	go func() {
		http.ListenAndServe("0.0.0.0:8899", nil)
	}()

	param := getParam()
	newServer(param)
}
