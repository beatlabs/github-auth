# GitHub Apps Authentication for Go
The `github-auth` package provide authentication support for GitHub Apps.

## Why?
The Go clients for GitHub do not handle authentication directly and an authenticated `*http.Client` is required.
The authentication is usually done using static tokens with `oauth2.StaticTokenSource()` which then provides an authenticated `*http.Client`.

With the introduction of GitHub Apps the authentication process requires JWT payloads.
This package provides an easy way to authenticate a Go application or service as a GitHub App (Installation).

The implementation is based on a slightly modified version of `golang.org/x/oauth2/jwt` to support GitHub JWT payloads and responses.

## How it works?
GitHub Apps use JWT for authentication.
The client can either authenticate as an App or as the App's Installation(s).
See [Authenticating with GitHub Apps](https://docs.github.com/en/free-pro-team@latest/developers/apps/authenticating-with-github-apps).

### Authentication as an App
JWT payloads are added to each request sent by the client.
See [Authenticating as a GitHub App](https://docs.github.com/en/free-pro-team@latest/developers/apps/authenticating-with-github-apps#authenticating-as-a-github-app)

### Authentication as an App's Installation
The client uses JWT as a token source and automatically requests temporary access tokens when required.
All requests are authenticated using the token.
See [Authenticating as an installation](https://docs.github.com/en/free-pro-team@latest/developers/apps/authenticating-with-github-apps#authenticating-as-an-installation)

By default all the repositories available to the installation are accessible by the token.
Optionally the access to repositories can be limited by either providing a list of repository IDs or names.

Also the access token's expiration can be specified.

## Requirements
1. A GitHub App. See [Creating a GitHub App](https://docs.github.com/en/free-pro-team@latest/developers/apps/creating-a-github-app).
2. The **App ID** which can be retrieved from GitHub (from the App's settings page or the API)
3. A **private key**. See [Generating a private key](https://docs.github.com/en/free-pro-team@latest/developers/apps/authenticating-with-github-apps#generating-a-private-key)
4. An **Installation ID** of the App's installed instance(s) (from Organization/repository installed Apps page or API):
    - See [Installing GitHub Apps in your organization](https://docs.github.com/en/free-pro-team@latest/github/customizing-your-github-workflow/installing-an-app-in-your-organization)
    - See [Installing GitHub Apps in your repository](https://docs.github.com/en/free-pro-team@latest/developers/apps/installing-github-apps)

## Usage
Install this module:
```shell
go get -u github.com/beatlabs/github-auth
```

To load the private key:
```go
import "github.com/beatlabs/github-auth/key"
...

// load from a file
key, err := key.FromFile("/path/to/file")

// load from data
key, err := key.Parse(bytes)
```

To authenticate as an App and get a client:
```go
import "github.com/beatlabs/github-auth/app"
...

// Create an App Config using the App ID and the private key
app, err := app.NewConfig(id, key)

// Get an *http.Client
client := app.Client()

// The client can be used to send authenticated requests
r, err := client.Get("https://api.github.com/app")
```

**Important:** when authenticating as an App, only specific API endpoints are accessible.
See [GitHub Apps REST API Reference](https://docs.github.com/en/free-pro-team@latest/rest/reference/apps) for the list of endpoints which support JWT.

To authenticate as an Installation:
```go
// Get the installation config from the authenticated App by providing the Installation ID
install, err := app.InstallationConfig(id)

// Or from scratch by providing the App ID, the private key and Installation ID
import "github.com/beatlabs/github-auth/app/inst"
...

install, err := inst.NewConfig(appID, installationID, key)

// Get an *http.Client
client = install.Client(ctx)


// The client can be used to send requests which are authenticated with temporary access tokens
r, err = client.Get("https://api.github.com/installation/repositories")
```

The returned `*http.Client` (App or Installation) can also be used to handle authentication for other Github clients.

The following client packages are tested:
- https://github.com/google/go-github for V3 (REST) API
- https://github.com/shurcooL/githubv4 for V4 (GraphQL) API

Using Google's `go-github`:
```go
client := github.NewClient(install.Client(ctx))
repos, _, err := client.Repositories.List(ctx, "", nil)
```

Using shurcooL's `githubv4`:
```go
client := githubv4.NewClient(install.Client(ctx))
...
err := client.Query(ctx, &query, nil)
```

### Enterprise
GitHub Enterprise App Installations are supported by using a custom URL:
```go
install , err := NewEnterpriseConfig(url, appID, installationID, key)
```
