package uaa

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// API is used to access a UAA server
type API interface {
	UserIDForAccountName(a string) (string, error)
}

// Authenticate will authenticate with a UAA server and set the Token and Client
// for the UAAAPI
func (a *Client) Authenticate() error {
	if a.Client != nil && a.Token != nil && a.Token.Valid() {
		return nil
	}

	config := oauth2.Config{
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth/authorize", a.URL),
			TokenURL: fmt.Sprintf("%s/oauth/token", a.URL),
		},
	}

	t, err := config.PasswordCredentialsToken(context.Background(), a.Username, a.Password)
	if err != nil {
		return errors.Wrap(err, "could not retrieve UAA token")
	}
	a.Token = t
	a.Client = config.Client(context.Background(), a.Token)
	return nil
}
