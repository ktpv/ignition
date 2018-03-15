package http

import (
	_ "expvar" // metrics
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pivotalservices/ignition/user"
	"golang.org/x/oauth2"
)

// API is the Ignition web app
type API struct {
	Domain       string
	Port         int
	ServePort    int
	WebRoot      string
	Scheme       string
	OAuth2Config *oauth2.Config
	Fetcher      user.Fetcher
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
	a.OAuth2Config.RedirectURL = fmt.Sprintf("%s%s", a.URI(), "/oauth2")
	r := a.createRouter()
	return http.ListenAndServe(fmt.Sprintf(":%v", a.ServePort), handlers.LoggingHandler(os.Stdout, handlers.CORS()(r)))
}

func (a *API) createRouter() *mux.Router {
	r := mux.NewRouter()
	r.Handle("/", ensureHTTPS(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, filepath.Join(a.WebRoot, "index.html"))
	}))).Name("index")
	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir(path.Join(a.WebRoot, "assets")+string(os.PathSeparator))))).Name("assets")
	r.Handle("/profile", ensureHTTPS(ContextFromSession(Authorize(profileHandler()))))
	a.handleAuth(r)
	r.Handle("/debug/vars", http.DefaultServeMux)
	return r
}
