package http

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dghubble/gologin"
	dgoauth2 "github.com/dghubble/gologin/oauth2"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// API is the Ignition web app
type API struct {
	Domain       string
	Port         int
	ServePort    int
	WebRoot      string
	AuthDomain   string
	Scheme       string
	clientID     string
	clientSecret string
}

// URI is the combination of the scheme, domain, and port
func (a *API) URI() string {
	s := fmt.Sprintf("%s://%s", a.Scheme, a.Domain)
	if a.Port != 0 {
		s = fmt.Sprintf("%s:%v", s, a.Port)
	}
	return s
}

// Run starts a server listening on the given serveURI
func (a *API) Run(clientID string, clientSecret string) error {
	err := a.validate(clientID, clientSecret)
	if err != nil {
		return err
	}

	r := a.createRouter()
	return http.ListenAndServe(fmt.Sprintf(":%v", a.ServePort), handlers.LoggingHandler(os.Stdout, handlers.CORS()(r)))
}

func (a *API) validate(clientID string, clientSecret string) error {
	a.clientID = clientID
	a.clientSecret = clientSecret
	if strings.TrimSpace(clientID) == "" {
		return errors.New("you must supply a non-empty client ID")
	}
	if strings.TrimSpace(clientSecret) == "" {
		return errors.New("you must supply a non-empty client secret")
	}
	return nil
}

func (a *API) createRouter() *mux.Router {
	r := mux.NewRouter()
	r.Handle("/", ensureHTTPS(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, filepath.Join(a.WebRoot, "index.html"))
	}))).Name("index")
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir(path.Join(a.WebRoot, "assets")+string(os.PathSeparator))))).Name("assets")
	c := oauth2.Config{
		ClientID:     a.clientID,
		ClientSecret: a.clientSecret,
		RedirectURL:  fmt.Sprintf("%s%s", a.URI(), "/oauth2"),
		Scopes:       []string{"user_attributes"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth/authorize", a.AuthDomain),
			TokenURL: fmt.Sprintf("%s/oauth/token", a.AuthDomain),
		},
	}
	stateConfig := gologin.DefaultCookieConfig
	r.Handle("/login", dgoauth2.StateHandler(stateConfig, dgoauth2.LoginHandler(&c, nil))).Name("login")
	return r
}
