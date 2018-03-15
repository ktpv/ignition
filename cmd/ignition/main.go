package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/pivotalservices/ignition/http"
	"github.com/pivotalservices/ignition/user"
	"github.com/pivotalservices/ignition/user/openid"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type config struct {
	port             int
	servePort        int
	domain           string
	webRoot          string
	scheme           string
	issuerURL        string
	clientID         string
	jwksURL          string
	authorizedDomain string
	oauth2Config     *oauth2.Config
	fetcher          user.Fetcher
}

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime | log.LUTC)
	c, err := buildConfig()
	if err != nil {
		log.Fatal(err)
	}

	api := http.API{
		WebRoot:          c.webRoot,
		Scheme:           c.scheme,
		Port:             c.port,
		Domain:           c.domain,
		ServePort:        c.servePort,
		OAuth2Config:     c.oauth2Config,
		AuthorizedDomain: c.authorizedDomain,
		Fetcher: &openid.Fetcher{
			Verifier: openid.NewVerifier(c.issuerURL, c.clientID, c.jwksURL),
		},
	}
	fmt.Println(fmt.Sprintf("Starting Server listening on %s", api.URI()))
	log.Fatal(api.Run())
}

func buildConfig() (*config, error) {
	c := &config{}
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	authVariant := os.Getenv("IGNITION_AUTH_VARIANT")
	issuerURL := os.Getenv("IGNITION_ISSUER_URL")
	authURL := os.Getenv("IGNITION_AUTH_URL")
	tokenURL := os.Getenv("IGNITION_TOKEN_URL")
	jwksURL := os.Getenv("IGNITION_JWKS_URL")
	clientID := os.Getenv("IGNITION_CLIENT_ID")
	clientSecret := os.Getenv("IGNITION_CLIENT_SECRET")

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
		switch strings.ToLower(strings.TrimSpace(authVariant)) {
		case "p-identity":
			service, err := env.Services.WithName("identity")
			if err != nil {
				return nil, errors.Wrap(err, "a Single Sign On service instance with the name \"identity\" is required to use this app")
			}
			clientid, ok := service.CredentialString("client_id")
			if !ok {
				return nil, errors.New("could not retrieve the client_id; make sure you have created and bound a Single Sign On service instance with the name \"identity\"")
			}
			clientID = clientid
			clientsecret, ok := service.CredentialString("client_secret")
			if !ok {
				return nil, errors.New("could not retrieve the client_secret; make sure you have created and bound a Single Sign On service instance with the name \"identity\"")
			}
			clientSecret = clientsecret
		}
	} else {
		c.scheme = "http"
		c.servePort = 3000
		c.port = 3000
		c.domain = "localhost"
		c.webRoot = filepath.Join(root, "web", "dist")
		c.scheme = "http"
	}

	authScopes := os.Getenv("IGNITION_AUTH_SCOPES")
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		Scopes: strings.Split(authScopes, ","),
	}
	c.oauth2Config = config
	c.jwksURL = jwksURL
	c.issuerURL = issuerURL
	c.clientID = clientID
	c.authorizedDomain = os.Getenv("IGNITION_AUTHORIZED_DOMAIN")
	return c, nil
}
