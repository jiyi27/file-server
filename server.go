package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type server struct {
	domain         string
	root           string
	rootAssetsPath string
	maxFileSize    int64
	sharedFiles    map[string]string
	auth           struct {
		username string
		password string
	}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// handle asset
	const assetPrefix = "asset="
	if strings.HasPrefix(r.URL.RawQuery, assetPrefix) {
		assetName := r.URL.RawQuery[len(assetPrefix):]
		s.asset(w, r, assetName)
		return
	}

	// handle shared file
	const sharedPrefix = "shared_id"
	if strings.HasPrefix(r.URL.RawQuery, sharedPrefix) {
		fileID := r.URL.RawQuery[len(sharedPrefix):]
		s.handleSharedDownload(w, r, fileID)
		return
	}

	// auth
	username, password, ok := r.BasicAuth()
	if ok {
		usernameHash := sha256.Sum256([]byte(username))
		passwordHash := sha256.Sum256([]byte(password))
		expectedUsernameHash := sha256.Sum256([]byte(s.auth.username))
		expectedPasswordHash := sha256.Sum256([]byte(s.auth.password))

		usernameMatch := subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1
		passwordMatch := subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1

		if usernameMatch && passwordMatch {
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

func (s *server) notifyAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="private", charset="UTF-8"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func (s *server) handleError(w http.ResponseWriter, _ *http.Request, status int, msg string) {
	msg = fmt.Sprintf("%v: %v", http.StatusText(status), msg)
	http.Error(w, msg, status)
}
