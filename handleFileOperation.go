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

func (s *server) handleMkdir(w http.ResponseWriter, r *http.Request, currentPath string) (error, int) {
	// Parse form.
	if err := r.ParseForm(); err != nil {
		err = fmt.Errorf("failed to parse folder name: %v", err)
		return err, http.StatusBadRequest
	}

	// Get username and password form the parsed form.
	folderName := r.Form.Get("folder_name")
	flPath := filepath.Join(currentPath, folderName)
	err := os.Mkdir(flPath, 0750)

	if err != nil {
		err = fmt.Errorf("failed to create folder: %v", err)
		return err, http.StatusBadRequest
	}

	// redirect
	r.URL.RawQuery = ""
	w.Header().Set("Location", r.URL.String())
	w.WriteHeader(http.StatusSeeOther)

	return nil, http.StatusOK
}

func (s *server) handleUpload(w http.ResponseWriter, r *http.Request, currentPath string) (error, int) {
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

	dstPath := filepath.Join(currentPath, filepath.Base(parsedFileHeader.Filename))
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

	// redirect
	r.URL.RawQuery = ""
	w.Header().Set("Location", r.URL.String())
	w.WriteHeader(http.StatusSeeOther)
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

	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir() && !files[j].IsDir() {
			return true // Directory should appear before the file
		} else if !files[i].IsDir() && files[j].IsDir() {
			return false // File should appear after the directory
		} else {
			// Both files[i] and files[j] are either directories or files
			return files[i].ModTime().After(files[j].ModTime())
		}
	})

	data := s.getTemplateData(r, files)
	tmpl, err := template.ParseFiles(s.rootAssetsPath + string(os.PathSeparator) + "index.html")
	if err != nil {
		return err
	}
	return tmpl.Execute(w, data)
}
