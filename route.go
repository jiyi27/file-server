package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (s *server) taskDelegation(w http.ResponseWriter, r *http.Request) {
	rawQuery := r.URL.RawQuery
	filePath := s.getFilePath(r)
	log.Printf(filePath)
	info, errStat := os.Stat(filePath)

	switch {
	case os.IsPermission(errStat):
		s.handleError(w, r, http.StatusForbidden, errStat.Error())
	case os.IsNotExist(errStat):
		s.handleError(w, r, http.StatusNotFound, errStat.Error())
	case errStat != nil:
		s.handleError(w, r, http.StatusInternalServerError, errStat.Error())
	case info.IsDir() && strings.HasPrefix(rawQuery, "upload"):
		err, status := s.handleUpload(w, r, filePath)
		if err != nil {
			s.handleError(w, r, status, err.Error())
		}
	case info.IsDir() && strings.HasPrefix(rawQuery, "mkdir"):
		err, status := s.handleMkdir(w, r, filePath)
		if err != nil {
			s.handleError(w, r, status, err.Error())
		}
	case strings.HasPrefix(rawQuery, "delete"):
		err := s.handleDelete(w, r, filePath)
		if err != nil {
			s.handleError(w, r, http.StatusInternalServerError, err.Error())
		}
	case info.IsDir():
		err := s.handleDir(w, r, filePath)
		if err != nil {
			log.Println(err)
			s.handleError(w, r, http.StatusInternalServerError, err.Error())
		}
	default:
		w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(info.Name()))
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeFile(w, r, filePath)
	}
}

func (s *server) getFilePath(r *http.Request) string {
	// all path format should be '/path/to/file' not '/path/to/file/'
	// separator is '\' on windows
	reqPath := strings.TrimSuffix(r.URL.Path, "/")
	reqPath = strings.ReplaceAll(reqPath, "/", string(os.PathSeparator))
	return filepath.Join(s.root, reqPath)
}
