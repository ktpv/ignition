package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/pivotalservices/ignition/http"
	"github.com/pkg/errors"
)

type config struct {
	port         int
	servePort    int
	domain       string
	webRoot      string
	scheme       string
	authVariant  string
	authURL      string
	tokenURL     string
	clientID     string
	clientSecret string
	authScopes   []string
}

func main() {
	c, err := buildConfig()
	if err != nil {
		log.Fatal(err)
	}

	api := http.API{
		WebRoot:     c.webRoot,
		Scheme:      c.scheme,
		Port:        c.port,
		Domain:      c.domain,
		ServePort:   c.servePort,
		AuthVariant: c.authVariant,
		AuthURL:     c.authURL,
		TokenURL:    c.tokenURL,
		AuthScopes:  c.authScopes,
	}
	fmt.Println(fmt.Sprintf("Starting Server listening on %s", api.URI()))
	log.Fatal(api.Run(c.clientID, c.clientSecret))
}

func buildConfig() (*config, error) {
	c := &config{}
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	authScopes := os.Getenv("IGNITION_AUTH_SCOPES")
	c.authScopes = strings.Split(authScopes, ",")

	if cfenv.IsRunningOnCF() {
		env, err := cfenv.Current()
		if err != nil {
			return nil, err
		}
		c.scheme = "https"
		c.servePort = env.Port
		c.port = 443
		if len(env.ApplicationURIs) == 0 {
			return nil, errors.New("ignition requires a route to function; please map a route")
		}
		c.domain = env.ApplicationURIs[0]
		c.webRoot = root
		c.authVariant = os.Getenv("IGNITION_AUTH_VARIANT")
		switch strings.ToLower(strings.TrimSpace(c.authVariant)) {
		case "p-identity":
			service, err := env.Services.WithName("identity")
			if err != nil {
				return nil, errors.Wrap(err, "a Single Sign On service instance with the name \"identity\" is required to use this app")
			}
			authDomain, ok := service.CredentialString("auth_domain")
			if !ok {
				return nil, errors.New("could not retrieve the auth_domain; make sure you have created and bound a Single Sign On service instance with the name \"identity\"")
			}
			c.authURL = fmt.Sprintf("%s/oauth/authorize", authDomain)
			c.tokenURL = fmt.Sprintf("%s/oauth/token", authDomain)
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
		default:
			c.authURL = os.Getenv("IGNITION_AUTH_URL")
			c.tokenURL = os.Getenv("IGNITION_TOKEN_URL")
			c.clientID = os.Getenv("IGNITION_CLIENT_ID")
			c.clientSecret = os.Getenv("IGNITION_CLIENT_SECRET")
		}
	} else {
		c.scheme = "http"
		c.servePort = 3000
		c.port = 3000
		c.domain = "localhost"
		c.webRoot = filepath.Join(root, "web", "dist")
		c.authVariant = os.Getenv("IGNITION_AUTH_VARIANT")
		c.authURL = os.Getenv("IGNITION_AUTH_URL")
		c.tokenURL = os.Getenv("IGNITION_TOKEN_URL")
		c.clientID = os.Getenv("IGNITION_CLIENT_ID")
		c.clientSecret = os.Getenv("IGNITION_CLIENT_SECRET")
		c.scheme = "http"
	}
	return c, nil
}
