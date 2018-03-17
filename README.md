## Ignition
[![CircleCI](https://circleci.com/gh/pivotalservices/ignition/tree/master.svg?style=svg)](https://circleci.com/gh/pivotalservices/ignition/tree/master)

A landing page for developers to self-service their way onto your Pivotal Cloud Foundry (PCF) deployment(s).

* Authenticates the user via OpenID Connect (which implicitly uses OAuth 2.0)
* Allows the user to access Apps Manager and view their personal PCF org

### Contribute

This application is a combination of a JavaScript single-page app (built with React) and a Go web app. The JavaScript app is built into a JavaScript bundle that the Go web app serves up. The Go web app also provides an API that the JavaScript app uses to function.

#### Yak Shaving (Developer Setup)

This project uses [`dep`](https://github.com/golang/dep) and [`yarn`](https://yarnpkg.com) for dependency management.

The following setup script shows how to get your MacOS workstation ready for `ignition` development. Don't just blindly execute shell scripts though; [take a thorough look through it](https://raw.githubusercontent.com/pivotalservices/ignition/master/setup.sh) and then run the following:

> `curl -o- https://raw.githubusercontent.com/pivotalservices/ignition/master/setup.sh | bash`

#### Add A Feature / Fix An Issue

We welcome pull requests to add additional functionality or fix issues. Please follow this procedure to get started:

1. Ensure you have `go` `>=1.10.x` and `node` `v8.x.x` installed
1. Ensure your `$GOPATH` is set; this is typically `$HOME/go`
1. Clone this repository: `go get -u github.com/pivotalservices/ignition`
1. Go to the repo root: `cd $GOPATH/src/github.com/pivotalservices/ignition`
1. *Fork this repository*
1. Add your fork as a new remote: `git remote add fork https://github.com/INSERT-YOUR-USERNAME-HERE/ignition.git`
1. Create a local branch: `git checkout -b your initials-your-feature-name` (e.g. `git checkout -b jf-add-logo`)
1. Make your changes, ensure you add tests to cover the changes, and then validate that all changes pass (see `Run all tests` below)
1. Push your feature branch to your fork: `git push fork your initials-your-feature-name` (e.g. `git push fork jf-add-logo`)
1. Make a pull request: `https://github.com/pivotalservices/ignition/compare/master...YOUR-USERNAME-HERE:your-initials-your-feature-name`

### Configure the application
#### Authentication
The app can be configured to authenticate against google or the PCF SSO tile.

To authenticate against google:
1. [Generate a goolge OAuth2 client id and secret](https://console.developers.google.com/apis/credentials)
1. Set the following environment variables
   * IGNITION_AUTH_SCOPES="openid,email,profile"
   * IGNITION_AUTH_URL="https://accounts.google.com/o/oauth2/v2/auth"
   * IGNITION_AUTH_VARIANT="google"
   * IGNITION_CLIENT_ID="[client id generated from google]"
   * IGNITION_CLIENT_SECRET="[client secret generated from google]"
   * IGNITION_TOKEN_URL="https://www.googleapis.com/oauth2/v4/token"
   * IGNITION_AUTH_ISSUER="https://accounts.google.com"
   * IGNITION_AUTHORIZED_DOMAIN="@pivotal.io"
   * IGNITION_JWKS_URL="https://www.googleapis.com/oauth2/v3/certs"
   * IGNITION_ISSUER_URL="https://accounts.google.com"

To authenticate against PCF SSO tile:
1. Configure the PCF SSO tile in your PCF foundation http://docs.pivotal.io/p-identity/
1. Set the following environment variables
   * IGNITION_AUTH_VARIANT: "p-identity"
   * IGNITION_ISSUER_URL: "https://ignition.uaa.run.pcfbeta.io/oauth/token"
   * IGNITION_AUTH_URL: "https://ignition.login.run.pcfbeta.io/oauth/authorize"
   * IGNITION_TOKEN_URL: "https://ignition.login.run.pcfbeta.io/oauth/token"
   * IGNITION_JWKS_URL: "https://ignition.login.run.pcfbeta.io/token_keys"
   * IGNITION_AUTH_SCOPES: "openid,profile,user_attributes"
   * IGNITION_AUTHORIZED_DOMAIN: "@pivotal.io"

### Run the application locally

1. Make sure you're in the repository root directory: `cd $GOPATH/src/github.com/pivotalservices/ignition`
1. Ensure the web bundle is built: `pushd web && yarn build && popd`
1. Start the go web app: `go run cmd/ignition/main.go`
1. Navigate to http://localhost:3000

### Run all tests

1. Make sure you're in the repository root directory: `cd $GOPATH/src/github.com/pivotalservices/ignition`
1. Run go tests: `go test ./...`
1. Run web tests: `pushd web && yarn ci && popd`

### Support

`ignition` is a community supported Pivotal Cloud Foundry add-on. [Opening an issue](https://github.com/pivotalservices/ignition/issues/new) for questions, feature requests and/or bugs is the best path to getting "support". We strive to be active in keeping this tool working and meeting your needs in a timely fashion.
