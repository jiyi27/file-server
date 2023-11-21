package main

import (
	"fmt"
	"net/http"
)

type user struct {
	username string
	password string
	path     string
}

func (u *user) notifyAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=\"%v\", charset=\"UTF-8\"", u.path))
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func (u *user) verifyAuth(password string) bool {
	return password == u.password
}
