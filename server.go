package main

import (
	"net/http"
	"os"
	"strings"
)

type server struct {
	root           string
	rootAssetsPath string
	maxFileSize    int64 // 1 << 25 == 32MB
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// handle asset
	const assetPrefix = "asset="
	if strings.HasPrefix(r.URL.RawQuery, assetPrefix) {
		assetName := r.URL.RawQuery[len(assetPrefix):]
		s.asset(w, r, assetName)
		return
	}

	// task delegation
	s.route(w, r)
}

func (s *server) asset(w http.ResponseWriter, r *http.Request, assetName string) {
	path := s.rootAssetsPath + string(os.PathSeparator) + assetName
	http.ServeFile(w, r, path)
}
