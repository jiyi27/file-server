package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *server) route(w http.ResponseWriter, r *http.Request) {
	// all path format should be '/path/to/file' not '/path/to/file/'
	// separator is '\' on windows
	reqPath := strings.TrimSuffix(r.URL.Path, "/")
	reqPath = strings.ReplaceAll(reqPath, "/", string(os.PathSeparator))
	osPath := filepath.Join(s.root, reqPath)
	info, errStat := os.Stat(osPath)
	switch {
	case os.IsPermission(errStat):
		_ = handleStatus(w, r, http.StatusForbidden)
	case os.IsNotExist(errStat):
		_ = handleStatus(w, r, http.StatusNotFound)
	case errStat != nil:
		_ = handleStatus(w, r, http.StatusInternalServerError)
	case info.IsDir() && r.Method == http.MethodPost:
		err, status := s.handleUpload(w, r, osPath)
		if err != nil {
			_ = handleStatus(w, r, status)
		}
		w.WriteHeader(status)
	case info.IsDir():
		err := s.handleDir(w, r, osPath)
		if err != nil {
			_ = handleStatus(w, r, http.StatusInternalServerError)
		}
	case !info.IsDir() && r.Method == http.MethodDelete:
		err := os.Remove(osPath)
		if err != nil {
			_ = handleStatus(w, r, http.StatusInternalServerError)
		}
	default:
		http.ServeFile(w, r, osPath)
	}
}

func handleStatus(w http.ResponseWriter, _ *http.Request, status int) error {
	w.WriteHeader(status)
	_, err := w.Write([]byte(http.StatusText(status)))
	if err != nil {
		return err
	}
	return nil
}
