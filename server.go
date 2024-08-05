package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func newServer(p *Param) {
	s := server{
		theme:       newTheme(),
		root:        p.root,
		maxFileSize: p.maxFileSize,
		files:       make(map[string]string),
		hsts:        p.Http != 0 && p.Https != 0,
		hstsMaxAge:  "31536000",
		toHttps:     p.Https != 0,
		authUsers:   p.users,
	}

	if err := s.init(); err != nil {
		log.Fatal(err)
	}

	switch {
	case p.Http == 0 && p.Https == 0: // no port specified, default http on 80
		serveHTTP(80, &s)
	case p.Http != 0 && p.Https == 0: // only plain http
		serveHTTP(p.Http, &s)
	case p.Http == 0 && p.Https != 0: // only https
		serveHTTPS(p.Https, p.SSLCert, p.SSLKey, &s)
	case p.Http != 0 && p.Https != 0: // both https and plain http
		go serveHTTP(p.Http, &s)
		serveHTTPS(p.Https, p.SSLCert, p.SSLKey, &s)
	}
}

type server struct {
	theme *Theme

	root        string
	maxFileSize int               // MB
	files       map[string]string // key: file id, value: file path

	hsts       bool // enable HSTS(HTTP Strict Transport Security).
	hstsMaxAge string
	toHttps    bool   // redirect plain HTTP request to HTTPS TLS port.
	authUsers  []user // username as index
}

func (s *server) init() error {
	// 750: -wxr-wr-----
	// x means can access directory.
	err := os.MkdirAll(s.root, 0750)
	if err != nil {
		return fmt.Errorf("faild to create root folder %v:%v", s.root, err)
	}

	// sort auth users by the depth of dir, the deepest dir depth at first.
	sort.Slice(s.authUsers, func(i, j int) bool {
		return pathDepth(s.authUsers[i].path) > pathDepth(s.authUsers[j].path)
	})

	// loop through all files in root dir, generate file id and save to map.
	err = filepath.Walk(s.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return filepath.SkipDir
		}
		if !info.IsDir() {
			s.files[generateHash(path)] = path
		}
		return nil
	})

	return nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("A new request from %v, method: %v, request path:%v", r.Host, r.Method, r.URL.Path)

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
	const idPrefix = "file_id="
	if strings.HasPrefix(r.URL.RawQuery, idPrefix) {
		fileID := r.URL.Query()[strings.TrimSuffix(idPrefix, "=")][0]
		s.handleSharedDownload(w, r, fileID)
		return
	}

	// format request path, empty path equals to "/".
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}

	// no auth users, handle request directly.
	if len(s.authUsers) == 0 {
		s.taskDelegation(w, r)
		return
	}

	authUser := s.retrieveUserByAuthPath(r.URL.Path)
	// no path needs auth, handle request directly.
	if authUser == nil {
		s.taskDelegation(w, r)
		return
	}

	username, password, ok := r.BasicAuth()
	if ok {
		if username == authUser.username && password == authUser.password {
			s.taskDelegation(w, r)
		} else {
			authUser.notifyAuth(w)
		}
		return
	}

	authUser.notifyAuth(w)
}

func (s *server) retrieveUserByAuthPath(path string) *user {
	for _, v := range s.authUsers {
		// the deepest dir depth match first.
		if strings.HasPrefix(path, v.path) {
			return &v
		}
	}
	return nil
}

func (s *server) asset(w http.ResponseWriter, r *http.Request, assetName string) {
	header := w.Header()
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Cache-Control", "public, max-age=3600")

	s.theme.RenderAsset(w, r, assetName)
}

func (s *server) handleError(w http.ResponseWriter, _ *http.Request, status int, msg string) {
	msg = fmt.Sprintf("%v: %v", http.StatusText(status), msg)
	http.Error(w, msg, status)
}
