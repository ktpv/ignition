package http

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type API struct {
	ServeURI     string
	WebRoot      string
	AuthDomain   string
	clientID     string
	clientSecret string
}

// Run starts a server listening on the given serveURI
func (a *API) Run(clientID string, clientSecret string) error {
	err := a.validate(clientID, clientSecret)
	if err != nil {
		return err
	}

	r := a.createRouter()
	return http.ListenAndServe(a.ServeURI, handlers.LoggingHandler(os.Stdout, handlers.CORS()(r)))
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
	return r
}
