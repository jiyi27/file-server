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

func newServer(p *Param) {
	toHttps := p.ListenTLS != 0
	// only available if server listens both plain HTTP and TLS on standard ports.
	hsts := p.ListenPlain != 0 && p.ListenTLS != 0

	s := server{
		root:           p.root,
		maxFileSize:    p.maxFileSize,
		rootAssetsPath: "template/",
		filesPathToId:  make(map[string]string),
		filesIdToPath:  make(map[string]string),
		hsts:           hsts,
		hstsMaxAge:     "31536000",
		toHttps:        toHttps,
		user:           p.user,
	}

	switch {
	case p.ListenPlain != 0 && p.ListenTLS == 0: // only plain http
		log.Printf("listening http on %v", p.ListenPlain)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", p.ListenPlain), &s))
	case p.ListenPlain == 0 && p.ListenTLS != 0: // only tls
		log.Printf("listening tls on %v", p.ListenTLS)
		log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%v", p.ListenTLS), p.TLSCert, p.TLSKey, &s))
	case p.ListenPlain != 0 && p.ListenTLS != 0: // both tls and plain http
		go func() {
			log.Printf("listening http on %v", p.ListenPlain)
			log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", p.ListenPlain), &s))
		}()
		log.Printf("listening tls on %v", p.ListenTLS)
		log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%v", p.ListenTLS), p.TLSCert, p.TLSKey, &s))
	default: // listen plain http on 80 port by default
		log.Printf("listening http on %s", ":8080")
		log.Fatal(http.ListenAndServe(":8080", &s))
	}
}

type server struct {
	root           string
	rootAssetsPath string // html and css files
	maxFileSize    int    // MB

	hsts       bool // enable HSTS(HTTP Strict Transport Security).
	hstsMaxAge string
	toHttps    bool // redirect plain HTTP request to HTTPS TLS port.

	filesPathToId map[string]string
	filesIdToPath map[string]string

	user struct {
		username string
		password string
	}
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// hsts redirect.
	if s.hsts && tryHsts(w, r) {
		return
	}

	// https redirect.
	if s.toHttps && tryToHttps(w, r) {
		return
	}

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
		fullPath := s.root + formatPath(filePath)

		url, err := s.generateSharedUrl(r, fullPath)
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
func (s *server) generateSharedUrl(r *http.Request, filePath string) (string, error) {
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
	}

	return fmt.Sprintf("%v://%v?shared_id=%v", getScheme(r), r.Host, id), nil
}

func (s *server) notifyAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="private", charset="UTF-8"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func (s *server) verifyAuth(username, password string) bool {
	usernameHash := sha256.Sum256([]byte(username))
	passwordHash := sha256.Sum256([]byte(password))
	expectedUsernameHash := sha256.Sum256([]byte(s.user.username))
	expectedPasswordHash := sha256.Sum256([]byte(s.user.password))

	usernameMatch := subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1
	passwordMatch := subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1

	return usernameMatch && passwordMatch
}

func (s *server) handleError(w http.ResponseWriter, _ *http.Request, status int, msg string) {
	msg = fmt.Sprintf("%v: %v", http.StatusText(status), msg)
	http.Error(w, msg, status)
}
