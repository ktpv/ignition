package uaa

import (
	"context"
	"fmt"

	"github.com/cloudfoundry-incubator/uaa-cli/uaa"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// API is used to access a UAA server
type API interface {
	UserIDForAccountName(a string) (string, error)
	CreateUser(username, origin, externalID, email string) (string, error)
}

// Authenticate will authenticate with a UAA server and set the Token and Client
// for the UAAAPI
func (a *Client) Authenticate() error {
	if a.Token == nil || a.Client == nil || !a.Token.Valid() {
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
	}

	if a.userManager == nil {
		uaaConfig := uaa.NewConfig()
		uaaConfig.AddTarget(uaa.Target{BaseUrl: a.URL})
		uaaConfig.AddContext(uaa.NewContextWithToken(a.Token.AccessToken))
		a.userManager = &uaa.UserManager{
			Config:     uaaConfig,
			HttpClient: a.Client,
		}
	}
	return nil
}
