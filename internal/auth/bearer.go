package auth

import (
	"context"
	"net/http"
)

// BearerAuth implements Bearer Token Authentication
type BearerAuth struct {
	token string
}

// NewBearerAuth creates a new Bearer authenticator
func NewBearerAuth(token string) *BearerAuth {
	return &BearerAuth{
		token: token,
	}
}

// Authenticate adds Bearer authentication to the request
func (a *BearerAuth) Authenticate(ctx context.Context, req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+a.token)
	return nil
}

// Type returns the auth type
func (a *BearerAuth) Type() string {
	return "bearer"
}
