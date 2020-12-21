package endpoint

import (
	"net/url"
)

var (
	// Default is the default GitHub api endpoint.
	Default = "https://api.github.com"
)

// Endpoint holds the GitHub API endpoint URL.
type Endpoint struct {
	url *url.URL
}

func new(raw string) (*Endpoint, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	return &Endpoint{url: u}, nil
}

// New returns a new endpoint with the default GitHub API URL.
func New() (*Endpoint, error) {
	return new(Default)
}

// NewEnterprise returns a new endpoint with the provided GitHub Enterprise API URL.
func NewEnterprise(url string) (*Endpoint, error) {
	return new(url)
}

// Get returns the full GitHub api endpoint for the provided uri.
func (e *Endpoint) Get(uri string) (string, error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return "", err
	}
	return e.url.ResolveReference(u).String(), nil
}
