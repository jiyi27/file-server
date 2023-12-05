package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (s *server) taskDelegation(w http.ResponseWriter, r *http.Request) {
	reqPath := formatPath(r.URL.Path)
	filePath := filepath.Join(s.root, reqPath)
	info, errStat := os.Stat(filePath)

	switch {
	case os.IsPermission(errStat):
		s.handleError(w, r, http.StatusForbidden, errStat.Error())
	case os.IsNotExist(errStat):
		s.handleError(w, r, http.StatusNotFound, errStat.Error())
	case errStat != nil:
		s.handleError(w, r, http.StatusInternalServerError, errStat.Error())
	case info.IsDir() && strings.HasPrefix(r.URL.RawQuery, "upload"):
		errs := s.handleUpload(w, r, filePath)
		if len(errs) == 0 {
			// clean url and redirect
			r.URL.RawQuery = ""
			w.Header().Set("Location", r.URL.String())
			w.WriteHeader(http.StatusSeeOther)
		} else {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(errs)
		}
	case info.IsDir() && strings.HasPrefix(r.URL.RawQuery, "mkdir"):
		err, status := s.handleMkdir(w, r, filePath)
		if err != nil {
			s.handleError(w, r, status, err.Error())
		}
	case strings.HasPrefix(r.URL.RawQuery, "delete"):
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
