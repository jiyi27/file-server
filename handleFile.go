package main

import (
	"fmt"
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

func (s *server) handleDir(w http.ResponseWriter, r *http.Request, currentDir string) error {
	d, err := os.Open(currentDir)
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
	return s.theme.RenderPage(w, data)
}

func (s *server) handleMkdir(w http.ResponseWriter, r *http.Request, currentDir string) (error, int) {
	// Parse form.
	if err := r.ParseForm(); err != nil {
		err = fmt.Errorf("failed to parse folder name: %v", err)
		return err, http.StatusBadRequest
	}

	// Get username and password form the parsed form.
	folderName := r.Form.Get("folder_name")
	flPath := filepath.Join(currentDir, folderName)
	// 750: -wxr-wr-----
	// x means can access directory.
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

func (s *server) handleUpload(w http.ResponseWriter, r *http.Request, currentDir string) (error, int) {
	var errs []error

	// limit the size of incoming request bodies.
	maxFileSize := int64(s.maxFileSize * 1024 * 1024)
	r.Body = http.MaxBytesReader(w, r.Body, maxFileSize)

	reader, err := r.MultipartReader()
	if err != nil {
		errs = append(errs, fmt.Errorf("an error occurred when parse requesr body:%v", err))
		return fmt.Errorf("an error occurred when parse uploaded file from requesr body:%v", err), http.StatusBadRequest
	}

	for {
		// reader.NextPart() will close the previous part automatically.
		part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF {
				break
			}
			errs = append(errs, fmt.Errorf("an error occurred when get file from multipart.Reader:%v", err))
			continue
		}

		// not a file, move to the next part.
		if part.FileName() == "" {
			continue
		}

		filename := getAvailableName(currentDir, part.FileName())
		dstPath := filepath.Join(currentDir, filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			errs = append(errs, fmt.Errorf("an error occurred when create file %v:%v", part.FileName(), err))
			continue
		}

		// io.Copy() will stream the file to dst, part is a reader with Read() method.
		_, err = io.Copy(dst, part)
		if err != nil {
			errs = append(errs, fmt.Errorf("an error occurred when copy file %v:%v", part.FileName(), err))
			continue
		}

		// considering the buffering mechanism, getting error when close a writable file is needed.
		if err = dst.Close(); err != nil {
			errs = append(errs, fmt.Errorf("an error occurred when close target file %v:%v", part.FileName(), err))
			continue
		}
	}

	// clean url and redirect
	r.URL.RawQuery = ""
	w.Header().Set("Location", r.URL.String())
	w.WriteHeader(http.StatusSeeOther)
	return nil, http.StatusSeeOther
}

func getAvailableName(fileDir, fileName string) string {
	filePath := path.Join(fileDir, fileName)
	// file already exists, generate a new name.
	if _, err := os.Stat(filePath); err == nil {
		fileName =
			strings.Split(fileName, ".")[0] +
				"_" +
				strconv.FormatInt(time.Now().UnixNano(), 10) +
				filepath.Ext(fileName)
	}
	// no such file, use the original name.
	return fileName
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
	filePath, ok := s.filesIdToPath[id]
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
