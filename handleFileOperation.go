package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

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

func (s *server) handleDir(w http.ResponseWriter, r *http.Request, osPath string) error {
	d, err := os.Open(osPath)
	if err != nil {
		return err
	}

	files, err := d.Readdir(-1)
	if err != nil {
		return err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })

	data := s.getTemplateData(r, files)
	tmpl, err := template.ParseFiles(templateFilePath)
	return tmpl.Execute(w, data)
}
