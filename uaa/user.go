package uaa

import (
	"net/http"
	"strings"

	"github.com/cloudfoundry-incubator/uaa-cli/uaa"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

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
