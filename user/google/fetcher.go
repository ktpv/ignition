package google

import (
	"context"

	"github.com/pivotalservices/ignition/user"
	"golang.org/x/oauth2"
	googleoauthv2 "google.golang.org/api/oauth2/v2"
)

// Fetcher retrieves the profile for a user when the auth variant is "google"
type Fetcher struct{}

// Profile retrieves the user's profile with the given context, config, and token
func (g *Fetcher) Profile(ctx context.Context, c *oauth2.Config, t *oauth2.Token) (*user.Profile, error) {
	httpClient := c.Client(ctx, t)
	googleService, err := googleoauthv2.New(httpClient)
	if err != nil {
		return nil, err
	}
	userInfoPlus, err := googleService.Userinfo.Get().Do()
	if err != nil {
		return nil, err
	}

	return &user.Profile{
		Email:       userInfoPlus.Email,
		AccountName: userInfoPlus.Email,
		Name:        userInfoPlus.Name,
	}, nil
}
