package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (s *server) handleDir(w http.ResponseWriter, r *http.Request, filePath string) error {
	d, err := os.Open(filePath)
	if err != nil {
		return err
	}

	files, err := d.Readdir(-1)
	if err != nil {
		return err
	}

	// make directory appear before the file
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir() && !files[j].IsDir() {
			return true
		} else if !files[i].IsDir() && files[j].IsDir() {
			return false
		} else {
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

	// clean url and redirect
	r.URL.RawQuery = ""
	w.Header().Set("Location", r.URL.String())
	w.WriteHeader(http.StatusSeeOther)

	return nil, http.StatusOK
}

func (s *server) handleUpload(w http.ResponseWriter, r *http.Request, currentPath string) (error, int) {
	// limit the size of incoming request bodies.
	r.Body = http.MaxBytesReader(w, r.Body, s.maxFileSize)
	// parse form from request body.
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

	// clean url and redirect
	r.URL.RawQuery = ""
	w.Header().Set("Location", r.URL.String())
	w.WriteHeader(http.StatusSeeOther)
	return nil, http.StatusSeeOther
}

func (s *server) handleDelete(w http.ResponseWriter, r *http.Request, filePath string) error {
	// remove removes the named file or (empty) directory.
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("faild to delete file:%v", err)
	}

	// clean url and redirect.
	// go back to the parent dir after delete file.
	r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
	r.URL.Path = path.Dir(r.URL.Path)
	r.URL.RawQuery = ""

	w.Header().Set("Location", r.URL.String())
	w.WriteHeader(http.StatusSeeOther)

	return nil
}

func (s *server) handleSharedDownload(w http.ResponseWriter, r *http.Request, id string) {
	fmt.Println(s.sharedFiles)
	filePath, ok := s.sharedFiles[id]
	if !ok {
		s.handleError(w, r, http.StatusNotFound, "no such file")
		return
	}

	info, err := os.Stat(filePath)
	if err != nil {
		s.handleError(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+strconv.Quote(info.Name()))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, filePath)
}

func (s *server) isEmpty(filePath string) (bool, error) {
	fileFullPath := path.Join(s.root, filePath)
	f, err := os.Open(fileFullPath)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}

	// either not empty or error
	return false, err
}
