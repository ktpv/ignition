package main

import (
	"fmt"
	"log"

	"github.com/pivotalservices/ignition/http"
)

func main() {
	serveURI := fmt.Sprintf(":%v", 3000)
	fmt.Println(fmt.Sprintf("Starting Server listening on %s", serveURI))
	log.Fatal(http.Run(serveURI))
}
