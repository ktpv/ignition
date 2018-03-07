package main

import (
	"fmt"
	"log"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/pivotalservices/ignition/http"
)

func main() {
	var serveURI string
	if cfenv.IsRunningOnCF() {
		env, err := cfenv.Current()
		if err != nil {
			log.Fatal(err)
		}
		serveURI = fmt.Sprintf(":%v", env.Port)
	} else {
		serveURI = fmt.Sprintf(":%v", 3000)
	}

	fmt.Println(fmt.Sprintf("Starting Server listening on %s", serveURI))
	log.Fatal(http.Run(serveURI))
}
