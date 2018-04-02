package session

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"golang.org/x/oauth2"
)

type key int

const (
	contextTokenKey key = iota
	contextUserIDKey
)

// TokenFromContext returns the Token from the ctx.
func TokenFromContext(ctx context.Context) (*oauth2.Token, error) {
	token, ok := ctx.Value(contextTokenKey).(*oauth2.Token)
	if !ok {
		return nil, errors.New("context missing Token")
	}
	return token, nil
}

// ContextWithToken returns a copy of ctx that stores the Token.
func ContextWithToken(ctx context.Context, token *oauth2.Token) context.Context {
	return context.WithValue(ctx, contextTokenKey, token)
}

// UserIDFromContext returns the user ID from the ctx.
func UserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(contextUserIDKey).(string)
	if !ok {
		return "", errors.New("context missing UserID")
	}
	return userID, nil
}

// ContextWithUserID returns a copy of ctx that stores the user ID.
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	if strings.TrimSpace(userID) == "" {
		return ctx
	}
	return context.WithValue(ctx, contextUserIDKey, userID)
}
