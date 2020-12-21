// Copyright 2021 Beat Research B.V. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package inst implements GitHub App Installation authentication.
//
// See: https://docs.github.com/en/free-pro-team@latest/developers/apps/authenticating-with-github-apps#authenticating-as-a-github-app
package inst

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"time"

	"github.com/beatlabs/github-auth/endpoint"
	"github.com/beatlabs/github-auth/jwt"
)

// Config defines an GitHub app installation config.
type Config struct {
	config jwt.Config
}

func new(endpoint endpoint.Endpoint, appID, instID string, key *rsa.PrivateKey) (*Config, error) {
	url, err := endpoint.Get(fmt.Sprintf("/app/installations/%s/access_tokens", instID))
	if err != nil {
		return nil, err
	}
	return &Config{
		config: jwt.Config{
			JWT:      jwt.JWT{AppID: appID, PrivateKey: key, Expires: time.Minute * 10},
			TokenURL: url,
		}}, nil
}

// NewConfig returns a new GitHub App instance.
func NewConfig(appID, instID string, key *rsa.PrivateKey) (*Config, error) {
	endpoint, err := endpoint.New()
	if err != nil {
		return nil, err
	}

	return new(*endpoint, appID, instID, key)
}

// NewEnterpriseConfig returns a new GitHub App instance.
func NewEnterpriseConfig(url, appID, instID string, key *rsa.PrivateKey) (*Config, error) {
	endpoint, err := endpoint.NewEnterprise(url)
	if err != nil {
		return nil, err
	}

	return new(*endpoint, appID, instID, key)
}

// SetRepositories returns an updated installation with the provided repositories.
// Access will be limited to the list of provided repositories
func (c *Config) SetRepositories(names []string) {
	c.config.Repositories.Names = names
}

// SetRepositoryIDs returns an updated installation with the provided repository ids.
// Access will be limited to the list of provided repository IDs.
//
func (c *Config) SetRepositoryIDs(ids []string) {
	c.config.Repositories.IDs = ids
}

// Client returns an HTTP client wrapping the context's
// HTTP transport and adding Authorization headers with tokens
// obtained using JWT.
//
// The returned client and its Transport should not be modified.
func (c *Config) Client(ctx context.Context) *http.Client {
	return c.config.Client(ctx)
}

// Permissions returns a map of the GitHub app client's permissions.
//
func (c *Config) Permissions() (map[string]string, error) {
	token, err := c.config.TokenSource(context.Background()).Token()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %v", err)
	}

	extra := token.Extra("permissions")
	pp, ok := extra.(map[string]string)
	if !ok {
		return nil, fmt.Errorf("failed to get permissions from extra field: %v", extra)
	}
	return pp, nil
}

// RepositorySelection returns the GitHub app client's repository selection (all or selected).
//
func (c *Config) RepositorySelection() (string, error) {
	token, err := c.config.TokenSource(context.Background()).Token()
	if err != nil {
		return "", fmt.Errorf("failed to get token: %v", err)
	}

	extra := token.Extra("repository_selection")
	rs, ok := extra.(string)
	if !ok {
		return "", fmt.Errorf("failed to get permissions from extra field: %v", extra)
	}
	return rs, nil
}
