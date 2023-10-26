package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type server struct {
	root           string
	rootAssetsPath string
	maxFileSize    int64
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
	s.taskDelegation(w, r)
}

func (s *server) asset(w http.ResponseWriter, r *http.Request, assetName string) {
	path := s.rootAssetsPath + string(os.PathSeparator) + assetName
	http.ServeFile(w, r, path)
}

func (s *server) taskDelegation(w http.ResponseWriter, r *http.Request) {
	rawQuery := r.URL.RawQuery
	currentPath := s.getCurrentPath(r)
	info, errStat := os.Stat(currentPath)

	switch {
	case os.IsPermission(errStat):
		handleError(w, r, http.StatusForbidden, errStat.Error())
	case os.IsNotExist(errStat):
		handleError(w, r, http.StatusNotFound, errStat.Error())
	case errStat != nil:
		handleError(w, r, http.StatusInternalServerError, errStat.Error())
	case info.IsDir() && strings.HasPrefix(rawQuery, "upload") && r.Method == http.MethodPost:
		err, status := s.handleUpload(w, r, currentPath)
		if err != nil {
			handleError(w, r, status, err.Error())
		}
	case info.IsDir() && strings.HasPrefix(rawQuery, "mkdir") && r.Method == http.MethodPost:
		err, status := s.handleMkdir(w, r, currentPath)
		if err != nil {
			handleError(w, r, status, err.Error())
		}
	case info.IsDir():
		err := s.handleDir(w, r, currentPath)
		if err != nil {
			handleError(w, r, http.StatusInternalServerError, err.Error())
		}
	// no request for directory below, all are files.
	case strings.HasPrefix(rawQuery, "delete"):
		err := s.handleDelete(w, r, currentPath)
		if err != nil {
			handleError(w, r, http.StatusInternalServerError, err.Error())
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

func handleError(w http.ResponseWriter, _ *http.Request, status int, msg string) {
	w.WriteHeader(status)
	msg = fmt.Sprintf("%v:%v", http.StatusText(status), msg)
	http.Error(w, msg, status)
	/*_, err := w.Write([]byte(msg))
	if err != nil {
		return err
	}
	return nil*/
}
