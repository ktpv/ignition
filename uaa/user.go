package uaa

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/cloudfoundry-incubator/uaa-cli/uaa"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// API is used to access a UAA server
type API interface {
	UserIDForAccountName(a string) (string, error)
}

// Client provides access to the UAA API
type Client struct {
	URL          string
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
	Token        *oauth2.Token
	Client       *http.Client
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

// UserIDForAccountName queries the UAA API for users filtered by account name
func (a *Client) UserIDForAccountName(accountName string) (string, error) {
	if strings.TrimSpace(accountName) == "" {
		return "", errors.New("cannot search for a user with an empty account name")
	}
	err := a.Authenticate()
	if err != nil {
		return "", errors.Wrap(err, "uaa: cannot authenticate")
	}

	ctx := uaa.NewContextWithToken(a.Token.AccessToken)
	config := uaa.NewConfig()
	config.AddTarget(uaa.Target{
		BaseUrl: a.URL,
	})
	config.AddContext(ctx)
	manager := uaa.UserManager{
		Config:     config,
		HttpClient: a.Client,
	}
	user, err := manager.GetByUsername(accountName, "", "")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(user.ID) == "" {
		return "", errors.Errorf("cannot find user with account name: [%s]", accountName)
	}
	return user.ID, nil
}
