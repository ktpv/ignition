package http

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// API is the Ignition web app
type API struct {
	Domain       string
	Port         int
	ServePort    int
	WebRoot      string
	Scheme       string
	AuthVariant  string
	AuthURL      string
	TokenURL     string
	AuthScopes   []string
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
	r.Handle("/profile", ensureHTTPS(ContextFromSession(Authorize(profileHandler()))))
	a.handleAuth(r)
	return r
}
