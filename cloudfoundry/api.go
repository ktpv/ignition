package cloudfoundry

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// API is a Cloud Foundry API endpoint
type API struct {
	URI          string
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
	Config       *oauth2.Config
	Token        *oauth2.Token
}

// URL returns the URL representation of the URI
func (a *API) URL() url.URL {
	u, err := url.Parse(a.URI)
	if err != nil || u == nil {
		log.Fatal(err)
	}
	return *u
}

// Login is a container for the login server JSON response
type login struct {
	Links struct {
		UAA   string `json:"uaa"`
		Login string `json:"login"`
	} `json:"links"`
}

func (a *API) ensureValidAuthentication() error {
	if a.Token == nil || !a.Token.Valid() {
		return a.Authenticate()
	}
	return nil
}

// Authenticate ensures that the API had a valid token that can be used for
// requests to the Cloud Controller API
func (a *API) Authenticate() error {
	i, err := a.Info()
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodGet, i.AuthorizationEndpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	var l login
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&l)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("could not decode response from URI: [%s]", i.AuthorizationEndpoint))
	}

	a.Config = &oauth2.Config{
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth/authorize", l.Links.Login),
			TokenURL: fmt.Sprintf("%s/oauth/token", l.Links.Login),
		},
	}
	t, err := a.Config.PasswordCredentialsToken(context.Background(), a.Username, a.Password)
	if err != nil {
		return errors.Wrap(err, "error retrieving UAA token")
	}
	a.Token = t
	return nil
}
