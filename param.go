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

	Http    int
	Https   int
	SSLKey  string
	SSLCert string
}

func (p *Param) init() {
	var users usersFlag

	flag.Var(&users, "auth", fmt.Sprintf("-auth <path:username:password>\n"+
		"example: /:admin:adminpw"))

	flag.StringVar(&p.root, "root", "./root", "Specify the root directory to save files")
	flag.StringVar(&p.root, "r", "./root", "(alias for -root)")
	flag.IntVar(&p.maxFileSize, "max", 32, "Specify maximum size for single file uploads in MB")

	flag.IntVar(&p.Http, "http", 0, "Enable HTTP server on the specified port, default is 80")
	flag.IntVar(&p.Https, "https", 0, "Enable HTTPS server on the specified port, default is 443")
	flag.StringVar(&p.SSLCert, "ssl-cert", "", "Specify the path of SSL certificate")
	flag.StringVar(&p.SSLKey, "ssl-key", "", "Specify the path of SSL private key")

	flag.Parse()

	// valid ssl
	if p.Https != 0 {
		if p.SSLCert == "" || p.SSLKey == "" {
			log.Fatal("HTTPS needs both ssl key and ssl cert to work")
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
