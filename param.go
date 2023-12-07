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

	users []user

	TLSKey      string
	TLSCert     string
	ListenPlain int
	ListenTLS   int
}

func (p *Param) init() {
	var users usersFlag

	flag.StringVar(&p.root, "root", "root/", "root directory for the file server")
	flag.StringVar(&p.root, "r", "root/", "(alias for -root)")
	flag.IntVar(&p.maxFileSize, "max", 32, "maximum size for single file uploads in MB")
	flag.IntVar(&p.ListenPlain, "plain-port", 0, "plain http port the server listens on")
	flag.IntVar(&p.ListenPlain, "p", 0, "(alias for -plain-port)")
	flag.IntVar(&p.ListenTLS, "tls-port", 0, "tls port the server listens on, will fail if cert or key is not specified")
	flag.StringVar(&p.TLSCert, "ssl-cert", "", "path to SSL server certificate")
	flag.StringVar(&p.TLSKey, "ssl-key", "", "path to SSL private key")
	flag.Var(&users, "auth", fmt.Sprintf("-auth <path:username:password>\n"+
		"specify user for HTTP Basic Auth"))

	flag.Parse()

	// valid tls
	if p.ListenTLS != 0 {
		if p.TLSCert == "" || p.TLSKey == "" {
			log.Fatal("if tls-port is specified, tls-key and tls-cert must be specified")
		}
	}

	p.users = users
}

type usersFlag []user

func (u *usersFlag) String() string {
	return fmt.Sprintf("%v", *u)
}

func (u *usersFlag) Set(value string) error {
	values := strings.Split(value, ":")
	if len(values) != 3 {
		return fmt.Errorf("wrong format of auth, format: path:username:password")
	}
	*u = append(
		*u,
		user{
			username: values[1],
			password: values[2],
			path:     formatPath(values[0]),
		},
	)
	return nil
}
