// Copyright 2021 Beat Research B.V. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package jwt implements GitHub JWT authentication.
//
// See: https://docs.github.com/en/free-pro-team@latest/developers/apps/authenticating-with-github-apps#authenticating-as-a-github-app
package jwt

import (
	"crypto/rsa"
	"net/http"
	"time"

	"github.com/beatlabs/github-auth/jws"
)

var (
	defaultHeader = &jws.Header{Algorithm: "RS256", Typ: "JWT"}
)

// JWT is the base structure for GitHub JWT.
type JWT struct {
	// AppID is the GitHub app ID.
	AppID string

	// PrivateKey contains the contents of an RSA private key or the
	// contents of a PEM file that contains a private key. The provided
	// private key is used to sign JWT payloads.
	// PEM containers with a passphrase are not supported.
	// Use the following command to convert a PKCS 12 file into a PEM.
	//
	//    $ openssl pkcs12 -in key.p12 -out key.pem -nodes
	//
	PrivateKey *rsa.PrivateKey

	// Expires optionally specifies how long the token is valid for.
	Expires time.Duration
}

// Payload returns the encoded GitHub JWT payload.
//
func (j *JWT) Payload() (string, error) {
	claimSet := &jws.ClaimSet{
		Iss: j.AppID,
	}
	if t := j.Expires; t > 0 {
		claimSet.Exp = time.Now().Add(t).Unix()
	}
	h := *defaultHeader
	payload, err := jws.Encode(&h, claimSet, j.PrivateKey)
	if err != nil {
		return "", err
	}

	return payload, nil
}

// Client returns an HTTP client wrapping the context's
// HTTP transport and adding Authorization headers.
//
func (j *JWT) Client() *http.Client {
	return &http.Client{
		Transport: &transport{j},
	}
}

// Custom transport for adding required HTTP headers.
//
type transport struct {
	jwt *JWT
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("Accept", "application/vnd.github.v3+json")
	payload, err := t.jwt.Payload()
	if err != nil {
		return nil, err
	}
	r.Header.Add("Authorization", "Bearer "+payload)
	return http.DefaultTransport.RoundTrip(r)
}
