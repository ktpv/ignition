package user

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
)

// Fetcher retrieves the profile for a user with the given token
type Fetcher interface {
	Profile(ctx context.Context, c *oauth2.Config, t *oauth2.Token) (*Profile, error)
}

// Profile is a user that can access ignition
type Profile struct {
	Email       string
	AccountName string
	Name        string
}

// unexported key type prevents collisions
type key int

const (
	profileKey key = iota
)

// WithProfile returns a copy of ctx that stores the Profile.
func WithProfile(ctx context.Context, profile *Profile) context.Context {
	return context.WithValue(ctx, profileKey, profile)
}

// ProfileFromContext returns the Profile from the ctx.
func ProfileFromContext(ctx context.Context) (*Profile, error) {
	profile, ok := ctx.Value(profileKey).(*Profile)
	if !ok {
		return nil, fmt.Errorf("Context missing Profile")
	}
	return profile, nil
}
