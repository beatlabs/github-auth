// Copyright 2014 The Go Authors. All rights reserved.
// Copyright 2021 Beat Research B.V.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package jwt implements the OAuth 2.0 JSON Web Token flow, commonly
// known as "two-legged OAuth 2.0".
// It has been modified to specifically support GitHub App JWT responses.
//
// See: https://tools.ietf.org/html/draft-ietf-oauth-jwt-bearer-12
// See: https://docs.github.com/en/free-pro-team@latest/rest/reference/apps#create-an-installation-access-token-for-an-app
package jwt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

// Config is the configuration for using GitHub JWT to fetch tokens.
type Config struct {
	JWT

	// Repositories is the list of repositories to limit the token access to.
	Repositories struct {
		// Names is the list of repository Names.
		Names []string `json:"repositories,omitempty"`

		// IDs is the list of repository IDs.
		IDs []string `json:"repository_ids,omitempty"`
	}

	// TokenURL is the GitHub App Installation URL for creating access tokens.
	// See: https://docs.github.com/en/free-pro-team@latest/rest/reference/apps#create-an-installation-access-token-for-an-app
	TokenURL string
}

// TokenSource returns a JWT TokenSource using the configuration
// in c and the HTTP client from the provided context.
func (c *Config) TokenSource(ctx context.Context) oauth2.TokenSource {
	return oauth2.ReuseTokenSource(nil, jwtSource{ctx, c})
}

// Client returns an HTTP client wrapping the context's
// HTTP transport and adding Authorization headers with tokens
// obtained from c.
//
// The returned client and its Transport should not be modified.
func (c *Config) Client(ctx context.Context) *http.Client {
	return oauth2.NewClient(ctx, c.TokenSource(ctx))
}

// jwtSource is a source that always does a signed JWT request for a token.
// It should typically be wrapped with a reuseTokenSource.
type jwtSource struct {
	ctx  context.Context
	conf *Config
}

func (js jwtSource) Token() (*oauth2.Token, error) {
	hc := oauth2.NewClient(js.ctx, nil)
	repos := new(bytes.Buffer)
	err := json.NewEncoder(repos).Encode(js.conf.Repositories)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, js.conf.TokenURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	payload, err := js.conf.Payload()
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+payload)
	resp, err := hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("oauth2: cannot fetch token: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("oauth2: cannot fetch token: %v", err)
	}
	if c := resp.StatusCode; c < 200 || c > 299 {
		return nil, &oauth2.RetrieveError{
			Response: resp,
			Body:     body,
		}
	}
	// tokenRes is the JSON response body.
	var tokenRes struct {
		AccessToken string `json:"token"`
		ExpiresAt   string `json:"expires_at"`
		TokenType   string
	}
	if err := json.Unmarshal(body, &tokenRes); err != nil {
		return nil, fmt.Errorf("oauth2: cannot fetch token: %v", err)
	}
	token := &oauth2.Token{
		AccessToken: tokenRes.AccessToken,
		TokenType:   "token",
	}
	raw := make(map[string]interface{})
	//nolint:errcheck
	json.Unmarshal(body, &raw) // no error checks for optional fields
	token = token.WithExtra(raw)

	if tokenRes.ExpiresAt != "" {
		token.Expiry, err = time.Parse(time.RFC3339, tokenRes.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("oauth2: cannot fetch token: %v", err)
		}
	}
	return token, nil
}
