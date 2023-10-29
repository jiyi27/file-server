package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type server struct {
	domain         string
	root           string
	rootAssetsPath string
	maxFileSize    int64
	filesPathToId  map[string]string
	filesIdToPath  map[string]string
	auth           struct {
		username string
		password string
	}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// handle asset.
	const assetPrefix = "asset="
	if strings.HasPrefix(r.URL.RawQuery, assetPrefix) {
		assetName := r.URL.Query()[strings.TrimSuffix(assetPrefix, "=")][0]
		s.asset(w, r, assetName)
		return
	}

	// handle download shared file.
	const idPrefix = "shared_id="
	if strings.HasPrefix(r.URL.RawQuery, idPrefix) {
		fileID := r.URL.Query()[strings.TrimSuffix(idPrefix, "=")][0]
		s.handleSharedDownload(w, r, fileID)
		return
	}

	// handle generate shared url of file.
	const pathPrefix = "filepath="
	if strings.HasPrefix(r.URL.RawQuery, pathPrefix) {
		filePath := r.URL.Query()[strings.TrimSuffix(pathPrefix, "=")][0]
		filePath = formatPath(filePath)
		fullPath := s.root + filePath

		url, err := s.generateSharedUrl(fullPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		if _, err = fmt.Fprint(w, url); err != nil {
			log.Printf("failed to send shared file link to client: %v", err)
			return
		}
		return
	}

	// handle auth.
	username, password, ok := r.BasicAuth()
	if ok {
		okVerify := s.verifyAuth(username, password)
		if okVerify {
			// task delegation
			s.taskDelegation(w, r)
			return
		}
	}

	s.notifyAuth(w)
}

func (s *server) asset(w http.ResponseWriter, r *http.Request, assetName string) {
	path := s.rootAssetsPath + string(os.PathSeparator) + assetName
	http.ServeFile(w, r, path)
}

// generateSharedUrl generates a link for the file and add its id and path to a map.
func (s *server) generateSharedUrl(filePath string) (string, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to generate shared url: %v", err)
	}

	if info.IsDir() {
		return "", errors.New("directory cannot be shared")
	}

	id, ok := s.filesPathToId[filePath]
	if !ok {
		id, err = s.generateID(10)
		if err != nil {
			return "", fmt.Errorf("failed to generate shared url: %v", err)
		}
		s.filesIdToPath[id] = filePath
		s.filesPathToId[filePath] = id
		s.domain = strings.TrimSuffix(s.domain, `/`)
		return fmt.Sprintf("%v/%v?shared_id=%v", s.domain, filePath, id), nil
	}

	return fmt.Sprintf("%v/%v?shared_id=%v", s.domain, filePath, id), nil
}

func (s *server) notifyAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="private", charset="UTF-8"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func (s *server) verifyAuth(username, password string) bool {
	usernameHash := sha256.Sum256([]byte(username))
	passwordHash := sha256.Sum256([]byte(password))
	expectedUsernameHash := sha256.Sum256([]byte(s.auth.username))
	expectedPasswordHash := sha256.Sum256([]byte(s.auth.password))

	usernameMatch := subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1
	passwordMatch := subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1

	return usernameMatch && passwordMatch
}

func (s *server) handleError(w http.ResponseWriter, _ *http.Request, status int, msg string) {
	msg = fmt.Sprintf("%v: %v", http.StatusText(status), msg)
	http.Error(w, msg, status)
}
