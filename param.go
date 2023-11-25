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
	var auths authFlags

	flag.StringVar(&p.root, "root", "root/", "root directory for the file server")
	flag.StringVar(&p.root, "r", "root/", "(alias for -root)")
	flag.IntVar(&p.maxFileSize, "max", 32, "maximum size for single file uploads in MB")
	flag.IntVar(&p.ListenPlain, "plain-port", 0, "plain http port the server listens on")
	flag.IntVar(&p.ListenPlain, "p", 0, "(alias for -plain-port)")
	flag.IntVar(&p.ListenTLS, "tls-port", 0, "tls port the server listens on, will fail if cert or key is not specified")
	flag.StringVar(&p.TLSCert, "ssl-cert", "", "path to SSL server certificate")
	flag.StringVar(&p.TLSKey, "ssl-key", "", "path to SSL private key")
	flag.Var(&auths, "auth", fmt.Sprintf("-auth <path:username:password>\n"+
		"specify user for HTTP Basic Auth"))

	flag.Parse()

	// valid tls
	if p.ListenTLS != 0 {
		if p.TLSCert == "" || p.TLSKey == "" {
			log.Fatal("if tls-port is specified, tls-key and tls-cert must be specified")
		}
	}

	for _, v := range auths {
		values := strings.Split(v, ":")
		if len(values) != 3 {
			log.Fatal("wrong format of auth, format: path:username:password")
		}
		p.users = append(
			p.users,
			user{
				username: values[1],
				password: values[2],
				path:     values[0],
			},
		)
	}
}

type authFlags []string

func (a *authFlags) String() string {
	return fmt.Sprintf("%v", *a)
}

func (a *authFlags) Set(value string) error {
	*a = append(*a, value)
	return nil
}
