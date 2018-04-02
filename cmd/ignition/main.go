package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/dghubble/sessions"
	"github.com/kelseyhightower/envconfig"
	"github.com/pivotalservices/ignition/http"
	"github.com/pivotalservices/ignition/uaa"
	"github.com/pivotalservices/ignition/user/openid"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type envConfig struct {
	AuthVariant       string   `envconfig:"auth_variant" default:"openid"`                        // IGNITION_AUTH_VARIANT
	ClientID          string   `envconfig:"client_id"`                                            // IGNITION_CLIENT_ID
	ClientSecret      string   `envconfig:"client_secret"`                                        // IGNITION_CLIENT_SECRET
	AuthURL           string   `envconfig:"auth_url" required:"true"`                             // IGNITION_AUTH_URL
	TokenURL          string   `envconfig:"token_url" required:"true"`                            // IGNITION_TOKEN_URL
	JWKSURL           string   `envconfig:"jwks_url" required:"true"`                             // IGNITION_JWKS_URL
	IssuerURL         string   `envconfig:"issuer_url" required:"true"`                           // IGNITION_ISSUER_URL
	AuthScopes        []string `envconfig:"auth_scopes" default:"openid,profile,user_attributes"` // IGNITION_AUTH_SCOPES
	AuthorizedDomain  string   `envconfig:"authorized_domain" required:"true"`                    // IGNITION_AUTHORIZED_DOMAIN
	SessionSecret     string   `envconfig:"session_secret" required:"true"`                       // IGNITION_SESSION_SECRET
	Port              int      `envconfig:"port" default:"3000"`                                  // IGNITION_PORT
	ServePort         int      `envconfig:"serve_port" default:"3000"`                            // IGNITION_SERVE_PORT
	Domain            string   `envconfig:"domain" default:"localhost"`                           // IGNITION_DOMAIN
	Scheme            string   `envconfig:"scheme" default:"http"`                                // IGNITION_SCHEME
	WebRoot           string   `envconfig:"web_root"`                                             // IGNITION_WEB_ROOT
	UAAURL            string   `envconfig:"uaa_url" required:"true"`                              // IGNITION_UAA_URL
	AppsURL           string   `envconfig:"apps_url" required:"true"`                             // IGNITION_APPS_URL
	CCAPIURL          string   `envconfig:"ccapi_url" required:"true"`                            // IGNITION_CCAPI_URL
	CCAPIClientID     string   `envconfig:"ccapi_client_id" default:"cf"`                         // IGNITION_CCAPI_CLIENT_ID
	CCAPIClientSecret string   `envconfig:"ccapi_client_secret" default:""`                       // IGNITION_CCAPI_CLIENT_SECRET
	CCAPIUsername     string   `envconfig:"ccapi_username" required:"true"`                       // IGNITION_CCAPI_USERNAME
	CCAPIPassword     string   `envconfig:"ccapi_password" required:"true"`                       // IGNITION_CCAPI_PASSWORD
	OrgPrefix         string   `envconfig:"org_prefix" default:"ignition"`                        // IGNITION_ORG_PREFIX
	QuotaID           string   `envconfig:"quota_id" required:"true"`                             // IGNITION_QUOTA_ID
}

func main() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime | log.LUTC)
	api, err := NewAPI()
	if err != nil {
		log.Fatal(err)
	}
	config := &cfclient.Config{
		ApiAddress: api.APIURL,
		Username:   api.APIUsername,
		Password:   api.APIPassword,
	}
	client, err := cfclient.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
	api.CCAPI = client
	log.Printf("Starting Server listening on %s\n", api.URI())
	log.Fatal(api.Run())
}

// NewAPI builds an http.API
func NewAPI() (*http.API, error) {
	var c envConfig
	err := envconfig.Process("ignition", &c)
	if err != nil {
		return nil, err
	}
	root, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(c.WebRoot) == "" {
		if cfenv.IsRunningOnCF() {
			c.WebRoot = root
		} else {
			c.WebRoot = filepath.Join(root, "web", "dist")
		}
	}

	env, err := cfenv.Current()
	if cfenv.IsRunningOnCF() && err != nil {
		return nil, err
	}

	if cfenv.IsRunningOnCF() {
		c.Scheme = "https"
		c.Port = 443
		c.ServePort = env.Port
		if len(env.ApplicationURIs) == 0 {
			return nil, errors.New("ignition requires a route to function; please map a route")
		}
		c.Domain = env.ApplicationURIs[0]
	}

	if cfenv.IsRunningOnCF() && strings.EqualFold(strings.TrimSpace(c.AuthVariant), "p-identity") {
		service, err := env.Services.WithName("identity")
		if err != nil {
			return nil, errors.Wrap(err, "a Single Sign On service instance with the name \"identity\" is required to use this app")
		}
		clientid, ok := service.CredentialString("client_id")
		if !ok {
			return nil, errors.New("could not retrieve the client_id; make sure you have created and bound a Single Sign On service instance with the name \"identity\"")
		}
		c.ClientID = clientid
		clientsecret, ok := service.CredentialString("client_secret")
		if !ok {
			return nil, errors.New("could not retrieve the client_secret; make sure you have created and bound a Single Sign On service instance with the name \"identity\"")
		}
		c.ClientSecret = clientsecret
	}

	if strings.TrimSpace(c.ClientID) == "" {
		return nil, errors.New("a client id must be set")
	}
	if strings.TrimSpace(c.ClientSecret) == "" {
		return nil, errors.New("a client secret must be set")
	}

	apiconfig := &oauth2.Config{
		ClientID:     c.CCAPIClientID,
		ClientSecret: c.CCAPIClientSecret,
		Scopes:       []string{"cloud_controller.admin"},
	}

	uaaAPI := &uaa.Client{
		URL:          c.UAAURL,
		ClientID:     c.CCAPIClientID,
		ClientSecret: c.CCAPIClientSecret,
		Username:     c.CCAPIUsername,
		Password:     c.CCAPIPassword,
	}

	api := http.API{
		WebRoot:   c.WebRoot,
		Scheme:    c.Scheme,
		Port:      c.Port,
		Domain:    c.Domain,
		ServePort: c.ServePort,
		AppsURL:   c.AppsURL,
		APIURL:    c.CCAPIURL,
		UAAURL:    c.UAAURL,
		UserConfig: &oauth2.Config{
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  c.AuthURL,
				TokenURL: c.TokenURL,
			},
			Scopes: c.AuthScopes,
		},
		APIConfig:        apiconfig,
		AuthorizedDomain: c.AuthorizedDomain,
		Fetcher: &openid.Fetcher{
			Verifier: openid.NewVerifier(c.IssuerURL, c.ClientID, c.JWKSURL),
		},
		SessionStore: sessions.NewCookieStore([]byte(c.SessionSecret), nil),
		APIUsername:  c.CCAPIUsername,
		APIPassword:  c.CCAPIPassword,
		OrgPrefix:    c.OrgPrefix,
		QuotaID:      c.QuotaID,
		UAAAPI:       uaaAPI,
	}
	return &api, nil
}
