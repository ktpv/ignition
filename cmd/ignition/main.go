package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/pivotalservices/ignition/http"
)

func main() {
	var serveURI string
	var webroot string
	root, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if cfenv.IsRunningOnCF() {
		env, err := cfenv.Current()
		if err != nil {
			log.Fatal(err)
		}
		serveURI = fmt.Sprintf(":%v", env.Port)
		webroot = root
	} else {
		serveURI = fmt.Sprintf(":%v", 3000)
		webroot = filepath.Join(root, "web", "dist")
	}

	fmt.Println(fmt.Sprintf("Starting Server listening on %s", serveURI))
	log.Fatal(http.Run(serveURI, webroot))
}
