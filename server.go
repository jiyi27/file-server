package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

func newServer(p *Param) {
	toHttps := p.ListenTLS != 0
	// only available if server listens both plain HTTP and TLS on standard ports.
	hsts := p.ListenPlain != 0 && p.ListenTLS != 0

	s := server{
		theme:         newTheme(),
		root:          p.root,
		maxFileSize:   p.maxFileSize,
		filesPathToId: make(map[string]string),
		filesIdToPath: make(map[string]string),
		hsts:          hsts,
		hstsMaxAge:    "31536000",
		toHttps:       toHttps,
		authUsers:     p.users,
	}

	// init server
	if err := s.init(); err != nil {
		log.Fatal(err)
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
	theme *Theme

	root        string
	maxFileSize int // MB

	hsts       bool // enable HSTS(HTTP Strict Transport Security).
	hstsMaxAge string
	toHttps    bool // redirect plain HTTP request to HTTPS TLS port.

	filesPathToId map[string]string // used for file sharing
	filesIdToPath map[string]string // used for file sharing

	authUsers []user // username as index
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
		return PathDepth(s.authUsers[i].path) > PathDepth(s.authUsers[j].path)
	})

	return nil
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

	// format request path, empty path equals to "/".
	if r.URL.Path == "" {
		r.URL.Path = "/"
	}

	// no auth users, handle request directly.
	if len(s.authUsers) == 0 {
		s.taskDelegation(w, r)
		return
	}

	authUser := s.getUserByPath(r.URL.Path)
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

func (s *server) getUserByPath(path string) *user {
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

func (s *server) handleError(w http.ResponseWriter, _ *http.Request, status int, msg string) {
	msg = fmt.Sprintf("%v: %v", http.StatusText(status), msg)
	http.Error(w, msg, status)
}
