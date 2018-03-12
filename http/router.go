package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dghubble/gologin"
	"github.com/dghubble/gologin/google"
	dgoauth2 "github.com/dghubble/gologin/oauth2"
	"github.com/dghubble/sessions"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

const (
	sessionTokenKey   = "token"
	sessionProfileKey = "profile"
	sessionEmailKey   = "email"
	sessionName       = "ignition"
	sessionSecret     = "cEM42gcY.rJaCmnZWay>hTXoAqYudMeY"
)

var sessionStore = sessions.NewCookieStore([]byte(sessionSecret), nil)

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
	c := oauth2.Config{
		ClientID:     a.clientID,
		ClientSecret: a.clientSecret,
		RedirectURL:  fmt.Sprintf("%s%s", a.URI(), "/oauth2"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  a.AuthURL,
			TokenURL: a.TokenURL,
		},
		Scopes: a.AuthScopes,
	}
	stateConfig := gologin.DefaultCookieConfig
	if a.Domain == "localhost" {
		stateConfig = gologin.DebugOnlyCookieConfig
	}
	r.Handle("/login", dgoauth2.StateHandler(stateConfig, dgoauth2.LoginHandler(&c, nil))).Name("login")
	r.Handle("/oauth2", dgoauth2.StateHandler(stateConfig, google.CallbackHandler(&c, IssueSession(), nil)))
	return r
}

// IssueSession stores the user's authentication state and profile in the
// session
func IssueSession() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		googleUser, err := google.UserFromContext(req.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		token, err := dgoauth2.TokenFromContext(req.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session := sessionStore.New(sessionName)
		j, err := json.Marshal(googleUser)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session.Values[sessionProfileKey] = string(j)
		j, err = json.Marshal(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session.Values[sessionTokenKey] = string(j)
		session.Save(w)
		http.Redirect(w, req, "/", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}
