package main

import (
	"net/http"
)

func tryHsts(w http.ResponseWriter, r *http.Request) (needRedirect bool) {
	if r.TLS != nil {
		w.Header().Set("Strict-Transport-Security", "max-age="+`31536000`)
		return
	}

	location := "https://" + r.Host + r.RequestURI
	http.Redirect(w, r, location, getRedirectCode(r))
	return true
}

func tryToHttps(w http.ResponseWriter, r *http.Request) (needRedirect bool) {
	if r.TLS != nil {
		return
	}

	location := "https://" + r.Host + r.RequestURI
	http.Redirect(w, r, location, getRedirectCode(r))
	return true
}

func getRedirectCode(r *http.Request) int {
	if r.Method == http.MethodPost {
		return http.StatusTemporaryRedirect
	} else {
		return http.StatusMovedPermanently
	}
}
