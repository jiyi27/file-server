package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func (s *server) route(w http.ResponseWriter, r *http.Request) {
	// all path format should be '/path/to/file' not '/path/to/file/'
	// separator is '\' on windows
	reqPath := strings.TrimSuffix(r.URL.Path, "/")
	reqPath = strings.ReplaceAll(reqPath, "/", string(filepath.Separator))
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
	case !info.IsDir() && r.Method == http.MethodDelete:
		err := os.Remove(osPath)
		if err != nil {
			_ = handleStatus(w, r, http.StatusInternalServerError)
		}
	case info.IsDir():
		err := serveDir(w, r, osPath)
		if err != nil {
			_ = serveStatus(w, r, http.StatusInternalServerError)
		}
	default:
		http.ServeFile(w, r, osPath)
	}
}

func (s *server) handleUpload(w http.ResponseWriter, r *http.Request, osPath string) (error, int) {
	// parse form from request body.
	r.Body = http.MaxBytesReader(w, r.Body, s.maxFileSize)
	if err := r.ParseMultipartForm(s.maxFileSize); err != nil {
		return fmt.Errorf("file is too large:%v", err), http.StatusBadRequest
	}

	// obtain file from parsed form.
	parsedFile, parsedFileHeader, err := r.FormFile("file")
	if errors.Is(err, http.ErrMissingFile) {
		w.Header().Set("Location", r.URL.String())
	}
	if err != nil {
		return err, http.StatusSeeOther
	}
	defer parsedFile.Close()

	// create root folder if it doesn't exist.
	err = os.MkdirAll(s.root, os.ModePerm)
	if err != nil {
		return err, http.StatusInternalServerError
	}

	dstPath := filepath.Join(osPath, filepath.Base(parsedFileHeader.Filename))
	var dst *os.File
	_, err = os.Stat(dstPath)
	// the name of parsed file already exists, create a dst file with a new name.
	if err == nil {
		filename := strings.Split(parsedFileHeader.Filename, ".")[0] +
			strconv.FormatInt(time.Now().UnixNano(), 10) +
			filepath.Ext(parsedFileHeader.Filename)
		dst, err = os.Create(filepath.Join(dstPath, filename))
	} else if os.IsNotExist(err) {
		// the name of parsed file already exists, create a dst file with original name.
		dst, err = os.Create(dstPath)
	}
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err), http.StatusInternalServerError
	}
	defer dst.Close()

	_, err = io.Copy(dst, parsedFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %v", err), http.StatusInternalServerError
	}

	// considering the buffering mechanism, getting error when close a writable file is needed.
	if err = dst.Close(); err != nil {
		return fmt.Errorf("failed to close dst fileHtml: %v", err), http.StatusInternalServerError
	}

	w.Header().Set("Location", r.URL.String())
	return nil, http.StatusSeeOther
}

func handleStatus(w http.ResponseWriter, _ *http.Request, status int) error {
	w.WriteHeader(status)
	_, err := w.Write([]byte(http.StatusText(status)))
	if err != nil {
		return err
	}
	return nil
}
