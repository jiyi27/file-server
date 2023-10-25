package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (s *server) route(w http.ResponseWriter, r *http.Request) {
	rawQuery := r.URL.RawQuery
	currentPath := s.getCurrentPath(r)
	info, errStat := os.Stat(currentPath)

	switch {
	case os.IsPermission(errStat):
		_ = handleStatus(w, r, http.StatusForbidden)
	case os.IsNotExist(errStat):
		_ = handleStatus(w, r, http.StatusNotFound)
	case errStat != nil:
		_ = handleStatus(w, r, http.StatusInternalServerError)
	case info.IsDir() && strings.HasPrefix(rawQuery, "upload") && r.Method == http.MethodPost:
		err, status := s.handleUpload(w, r, currentPath)
		if err != nil {
			_ = handleStatus(w, r, status)
			log.Println(err)
		}
	case info.IsDir() && strings.HasPrefix(rawQuery, "mkdir") && r.Method == http.MethodPost:
		err, status := s.handleMkdir(w, r, currentPath)
		if err != nil {
			_ = handleStatus(w, r, status)
			log.Println(err)
		}
	case info.IsDir():
		err := s.handleDir(w, r, currentPath)
		if err != nil {
			_ = handleStatus(w, r, http.StatusInternalServerError)
			log.Println(err)
		}
	case strings.HasPrefix(rawQuery, "download"):

	case info.IsDir():
		err := s.handleDir(w, r, currentPath)
		if err != nil {
			_ = handleStatus(w, r, http.StatusInternalServerError)
			log.Println(err)
		}
	case !info.IsDir() && r.Method == http.MethodDelete:
		err := os.Remove(currentPath)
		if err != nil {
			_ = handleStatus(w, r, http.StatusInternalServerError)
			log.Println(err)
		}
	default:
		w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(info.Name()))
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeFile(w, r, currentPath)
	}
}

func (s *server) getCurrentPath(r *http.Request) string {
	// all path format should be '/path/to/file' not '/path/to/file/'
	// separator is '\' on windows
	reqPath := strings.TrimSuffix(r.URL.Path, "/")
	reqPath = strings.ReplaceAll(reqPath, "/", string(os.PathSeparator))
	return filepath.Join(s.root, reqPath)
}

func handleStatus(w http.ResponseWriter, _ *http.Request, status int) error {
	w.WriteHeader(status)
	_, err := w.Write([]byte(http.StatusText(status)))
	if err != nil {
		return err
	}
	return nil
}
