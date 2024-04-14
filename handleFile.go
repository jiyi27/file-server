package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func (s *server) handleDir(w http.ResponseWriter, r *http.Request, currentDir string) error {
	d, err := os.Open(currentDir)
	if err != nil {
		return err
	}
	defer d.Close()

	files, err := d.Readdir(-1)
	if err != nil {
		return err
	}

	// sort files, directories first, then sort by modified time.
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

func (s *server) handleUpload(_ http.ResponseWriter, r *http.Request, currentDir string) (errs []uploadError) {
	// limit the size of incoming request bodies.
	r.Body = &LimitedReader{r: r.Body, n: int64(s.maxFileSize * 1024 * 1024)}

	reader, err := r.MultipartReader()
	if err != nil {
		errs = append(errs, uploadError{Message: fmt.Sprintf("parse request body: %v", err)})
		return
	}

	for {
		// reader.NextPart() will close the previous part automatically.
		part, err := reader.NextPart()
		if err != nil {
			if err == io.EOF { // finish reading all parts, exit the loop
				break
			}

			errs = append(errs, uploadError{Message: fmt.Sprintf("reader.NextPart(): %v", err)})

			// too many errors, stop uploading, you must limit the number of errors in case of infinite loop.
			if len(errs) >= 10 {
				errs = append(errs, uploadError{Message: "Maximum error limit reached"})
				return
			}

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
			errs = append(errs, uploadError{
				FileName: part.FileName(),
				Message:  fmt.Sprintf("create file: %v", err),
			})
			continue
		}

		// io.Copy() will stream the file to dst, multipart.Part is an io.Reader.
		_, err = io.Copy(dst, part)
		if err != nil {
			errs = append(errs, uploadError{
				FileName: part.FileName(),
				Message:  fmt.Sprintf("copy file: %v", err),
			})
			_ = dst.Close()

			// copy failed, delete the file
			_ = os.Remove(dstPath)

			// file too large, stop uploading
			var er *FileToLarge
			if errors.As(err, &er) {
				return
			}
			continue
		}

		// closing a writer may flush all buffered data to the underlying io.Writer.
		// this fails means the file is not written completely.
		// so we need to delete the file.
		if err = dst.Close(); err != nil {
			errs = append(errs, uploadError{
				FileName: part.FileName(),
				Message:  fmt.Sprintf("close file: %v", err),
			})

			// close failed, delete the file
			_ = os.Remove(dstPath)

			continue
		}

		// add file to the map
		id := generateHash(dstPath)
		s.files[id] = dstPath
	}

	return
}

func (s *server) handleUploadLargeFile(_ http.ResponseWriter, r *http.Request, currentDir string) *uploadError {
	filename := r.Header.Get("X-Filename")
	if filename == "" {
		return &uploadError{Message: "Filename is required in X-Filename header"}
	}

	filename = getAvailableName(currentDir, filename)

	dstPath := filepath.Join(currentDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return &uploadError{Message: fmt.Sprintf("create file: %v", err)}
	}

	_, err = io.Copy(dst, r.Body)
	if err != nil {
		_ = dst.Close()
		_ = os.Remove(dstPath)
		return &uploadError{Message: fmt.Sprintf("copy file: %v", err)}
	}

	id := generateHash(dstPath)
	s.files[id] = dstPath
	return nil
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
	log.Println("download shared file:", id)
	// loop through s.files and print out the key and value.
	for k, v := range s.files {
		log.Printf("key[%s] value[%s]\n", k, v)
	}

	filePath, ok := s.files[id]
	if !ok {
		s.handleError(w, r, http.StatusNotFound, "download shared file: no such file")
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

type uploadError struct {
	FileName string `json:"fileName"`
	Message  string `json:"message"`
}
