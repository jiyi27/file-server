package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

func getParam() *Param {
	p := Param{}
	p.init()
	return &p
}

type Param struct {
	root        string
	assetPath   string
	maxFileSize int // MB

	user struct {
		username string
		password string
	}

	TLSKey      string
	TLSCert     string
	ListenPlain int
	ListenTLS   int
}

func (p *Param) init() {
	user := "admin:admin"

	flag.StringVar(&p.root, "root", "root/", "root directory for the file server")
	flag.StringVar(&p.root, "r", "root/", "(alias for -root)")
	flag.StringVar(&p.assetPath, "asset", "template/", "directory for storing assets (html, css files)")
	flag.IntVar(&p.maxFileSize, "max", 32, "maximum size for single file uploads in MB")
	flag.IntVar(&p.ListenPlain, "plain-port", 0, "plain http port the server listens on")
	flag.IntVar(&p.ListenPlain, "p", 0, "(alias for -plain-port)")
	flag.IntVar(&p.ListenTLS, "tls-port", 0, "tls port the server listens on, will fail if cert or key is not specified")
	flag.StringVar(&p.TLSCert, "ssl-cert", "", "path to SSL server certificate")
	flag.StringVar(&p.TLSKey, "ssl-key", "", "path to SSL private key")
	flag.StringVar(&user, "user", user, fmt.Sprintf("--user <username:password>\n"+
		"specify user for HTTP Basic Auth"))

	flag.Parse()

	// valid tls
	if p.ListenTLS != 0 {
		if p.TLSCert == "" || p.TLSKey == "" {
			log.Fatal("if tls-port is specified, tls-key and tls-cert must be specified")
		}
	}

	// parse and valid user
	index := strings.Index(user, ":")
	p.user.username = user[:index]
	p.user.password = user[index+1:]
	if p.user.username == "" || p.user.password == "" {
		log.Fatal("username or password is empty, format: \n   username:password")
	}
}
