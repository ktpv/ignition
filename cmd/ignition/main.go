package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/pivotalservices/ignition/http"
	"github.com/pkg/errors"
)

type config struct {
	serveURI     string
	webRoot      string
	authDomain   string
	clientID     string
	clientSecret string
}

func main() {
	c, err := buildConfig()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("Starting Server listening on %s", c.serveURI))
	api := http.API{
		ServeURI:   c.serveURI,
		WebRoot:    c.webRoot,
		AuthDomain: c.authDomain,
	}
	log.Fatal(api.Run(c.clientID, c.clientSecret))
}

func buildConfig() (*config, error) {
	c := &config{}
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if cfenv.IsRunningOnCF() {
		env, err := cfenv.Current()
		if err != nil {
			return nil, err
		}
		c.serveURI = fmt.Sprintf(":%v", env.Port)
		c.webRoot = root
		service, err := env.Services.WithName("identity")
		if err != nil {
			return nil, errors.Wrap(err, "a Single Sign On service instance with the name \"identity\" is required to use this app")
		}
		authDomain, ok := service.CredentialString("auth_domain")
		if !ok {
			return nil, errors.New("could not retrieve the auth_domain; make sure you have created and bound a Single Sign On service instance with the name \"identity\"")
		}
		c.authDomain = authDomain
		clientID, ok := service.CredentialString("client_id")
		if !ok {
			return nil, errors.New("could not retrieve the client_id; make sure you have created and bound a Single Sign On service instance with the name \"identity\"")
		}
		c.clientID = clientID
		clientSecret, ok := service.CredentialString("client_secret")
		if !ok {
			return nil, errors.New("could not retrieve the client_secret; make sure you have created and bound a Single Sign On service instance with the name \"identity\"")
		}
		c.clientSecret = clientSecret
	} else {
		c.serveURI = fmt.Sprintf(":%v", 3000)
		c.webRoot = filepath.Join(root, "web", "dist")
		c.authDomain = os.Getenv("IGNITION_AUTH_DOMAIN")
		c.clientID = os.Getenv("IGNITION_CLIENT_ID")
		c.clientSecret = os.Getenv("IGNITION_CLIENT_SECRET")
	}
	return c, nil
}
