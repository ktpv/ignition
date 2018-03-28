package http

import (
	_ "expvar" // metrics
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/dghubble/sessions"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pivotalservices/ignition/user"
	"golang.org/x/oauth2"
)

// API is the Ignition web app
type API struct {
	AuthorizedDomain string
	SessionSecret    string
	Domain           string
	Port             int
	ServePort        int
	WebRoot          string
	Scheme           string
	APIURL           string
	AppsURL          string
	UAAURL           string
	UserConfig       *oauth2.Config
	APIConfig        *oauth2.Config
	APIUsername      string
	APIPassword      string
	Fetcher          user.Fetcher
	SessionStore     sessions.Store
	CCAPI            *cfclient.Client
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
func (a *API) Run() error {
	a.UserConfig.RedirectURL = fmt.Sprintf("%s%s", a.URI(), "/oauth2")
	r := a.createRouter()
	return http.ListenAndServe(fmt.Sprintf(":%v", a.ServePort), handlers.LoggingHandler(os.Stdout, handlers.CORS()(r)))
}

func (a *API) createRouter() *mux.Router {
	r := mux.NewRouter()
	r.Handle("/", ensureHTTPS(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, filepath.Join(a.WebRoot, "index.html"))
	}))).Name("index")
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir(path.Join(a.WebRoot, "assets")+string(os.PathSeparator))))).Name("assets")
	r.Handle("/profile", ensureHTTPS(a.ContextFromSession(Authorize(profileHandler(), a.AuthorizedDomain))))
	r.Handle("/organization", ensureHTTPS(a.ContextFromSession(Authorize(organizationHandler(a.AppsURL, a.CCAPI), a.AuthorizedDomain))))
	a.handleAuth(r)
	r.HandleFunc("/403", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	})
	r.Handle("/debug/vars", http.DefaultServeMux)
	return r
}
