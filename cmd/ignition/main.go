package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-community/go-cfenv"
	oidc "github.com/coreos/go-oidc"
	"github.com/pivotalservices/ignition/http"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type config struct {
	port         int
	servePort    int
	domain       string
	webRoot      string
	scheme       string
	oauth2Config *oauth2.Config
}

func main() {

	c, err := buildConfig(oidcEndpoint)
	if err != nil {
		log.Fatal(err)
	}

	api := http.API{
		WebRoot:      c.webRoot,
		Scheme:       c.scheme,
		Port:         c.port,
		Domain:       c.domain,
		ServePort:    c.servePort,
		OAuth2Config: c.oauth2Config,
	}
	fmt.Println(fmt.Sprintf("Starting Server listening on %s", api.URI()))
	log.Fatal(api.Run())
}

func oidcEndpoint(ctx context.Context, issuer string) (*oauth2.Endpoint, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}
	endpoint := provider.Endpoint()
	return &endpoint, nil
}

func buildConfig(endpointFunc func(ctx context.Context, issuer string) (*oauth2.Endpoint, error)) (*config, error) {
	c := &config{}
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	authVariant := os.Getenv("IGNITION_AUTH_VARIANT")
	authIssuer := os.Getenv("IGNITION_AUTH_ISSUER")
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
			authDomain, ok := service.CredentialString("auth_domain")
			if !ok {
				return nil, errors.New("could not retrieve the auth_domain; make sure you have created and bound a Single Sign On service instance with the name \"identity\"")
			}
			authIssuer = authDomain
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
	endpoint, err := endpointFunc(context.Background(), authIssuer)
	if err != nil {
		log.Fatal(err)
	}
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     *endpoint,
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	c.oauth2Config = config
	return c, nil
}
