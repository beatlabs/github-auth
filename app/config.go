// Copyright 2021 Beat Research B.V. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package app implements GitHub App authentication.
// See: https://docs.github.com/en/free-pro-team@latest/developers/apps/authenticating-with-github-apps#authenticating-as-a-github-app
package app

import (
	"crypto/rsa"
	"net/http"
	"time"

	"github.com/beatlabs/github-auth/app/inst"
	"github.com/beatlabs/github-auth/jwt"
)

// Config defines the base GitHub App Config structure.
type Config struct {
	jwt jwt.JWT
}

// NewConfig returns a new GitHub App instance.
func NewConfig(id string, key *rsa.PrivateKey) (*Config, error) {
	return &Config{jwt: jwt.JWT{AppID: id, PrivateKey: key, Expires: time.Minute * 10}}, nil
}

// Client returns an HTTP client with an HTTP transport that adds Authorization headers.
//
func (c *Config) Client() *http.Client {
	return c.jwt.Client()
}

// InstallationConfig returns the Installation Config for the provided installation ID.
func (c *Config) InstallationConfig(id string) (*inst.Config, error) {
	return inst.NewConfig(c.jwt.AppID, id, c.jwt.PrivateKey)
}
