package openid

import (
	"context"
	"fmt"

	oidc "github.com/coreos/go-oidc"
	"github.com/pivotalservices/ignition/user"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// Fetcher retrieves the profile for a user when the auth variant is "google"
type Fetcher struct {
	Verifier *oidc.IDTokenVerifier
}

// Profile retrieves the user's profile with the given context, config, and token
func (g *Fetcher) Profile(ctx context.Context, c *oauth2.Config, t *oauth2.Token) (*user.Profile, error) {
	if g.Verifier == nil {
		return nil, errors.New("unable to verify token")
	}
	rawIDToken, ok := t.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("profile: no id_token")
	}
	idToken, err := g.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, errors.Wrap(err, "unable to verify rawIDToken")
	}

	var claims struct {
		Sub        string `json:"sub"`
		UserName   string `json:"user_name"`
		UserID     string `json:"user_id"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
		Email      string `json:"email"`
	}
	if err = idToken.Claims(&claims); err != nil {
		return nil, err
	}

	return &user.Profile{
		Email:       claims.Email,
		AccountName: claims.UserName,
		Name:        fmt.Sprintf("%s %s", claims.GivenName, claims.FamilyName),
	}, nil
}

// NewVerifier returns an IDTokenVerifier that uses a keySet fetched from jwksURL
func NewVerifier(issuerURL string, clientID string, jwksURL string) *oidc.IDTokenVerifier {
	keySet := oidc.NewRemoteKeySet(context.Background(), jwksURL)
	return oidc.NewVerifier(issuerURL, keySet, &oidc.Config{
		ClientID: clientID,
	})
}
