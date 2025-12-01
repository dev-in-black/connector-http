package auth

import (
	"context"
	"net/http"
)

// BasicAuth implements HTTP Basic Authentication
type BasicAuth struct {
	username string
	password string
}

// NewBasicAuth creates a new Basic authenticator
func NewBasicAuth(username, password string) *BasicAuth {
	return &BasicAuth{
		username: username,
		password: password,
	}
}

// Authenticate adds Basic authentication to the request
func (a *BasicAuth) Authenticate(ctx context.Context, req *http.Request) error {
	req.SetBasicAuth(a.username, a.password)
	return nil
}

// Type returns the auth type
func (a *BasicAuth) Type() string {
	return "basic"
}
