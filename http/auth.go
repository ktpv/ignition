package http

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dghubble/gologin"
	dgoauth2 "github.com/dghubble/gologin/oauth2"
	"github.com/dghubble/sessions"
	"github.com/gorilla/mux"
	"github.com/pivotalservices/ignition/user"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type key int

const (
	sessionTokenKey       = "token"
	sessionProfileKey     = "profile"
	sessionEmailKey       = "email"
	sessionName           = "ignition"
	sessionSecret         = "cEM42gcY.rJaCmnZWay>hTXoAqYudMeY"
	contextTokenKey   key = iota
)

var sessionStore = sessions.NewCookieStore([]byte(sessionSecret), nil)

func (a *API) handleAuth(r *mux.Router) {
	stateConfig := gologin.DefaultCookieConfig
	if a.Domain == "localhost" {
		stateConfig = gologin.DebugOnlyCookieConfig
	}
	r.Handle("/login", ensureHTTPS(dgoauth2.StateHandler(stateConfig, dgoauth2.LoginHandler(a.OAuth2Config, nil)))).Name("login")
	r.Handle("/oauth2", ensureHTTPS(dgoauth2.StateHandler(stateConfig, CallbackHandler(a.OAuth2Config, a.Fetcher, IssueSession(), http.HandlerFunc(LogoutHandler))))).Name("oauth2")
	r.Handle("/logout", ensureHTTPS(http.HandlerFunc(LogoutHandler))).Name("logout")
}

// Authorize guards access to protected resources by inspecting the user's token
func Authorize(next http.Handler, domain string) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		token, ok := req.Context().Value(contextTokenKey).(*oauth2.Token)
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		if token.Expiry.UTC().Sub(time.Now().UTC()) < 0 {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		profile, err := user.ProfileFromContext(req.Context())
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		if strings.TrimSpace(domain) != "" && !strings.HasSuffix(strings.ToLower(profile.Email), domain) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, req)
	}
	return http.HandlerFunc(fn)
}

// Write gzipped data to a Writer
func gzipWrite(w io.Writer, data []byte) error {
	// Write gzipped data to the client
	gw, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	defer gw.Close()
	gw.Write(data)
	return err
}

// Write gunzipped data to a Writer
func gunzipWrite(w io.Writer, data []byte) error {
	// Write gzipped data to the client
	gr, err := gzip.NewReader(bytes.NewBuffer(data))
	defer gr.Close()
	data, err = ioutil.ReadAll(gr)
	if err != nil {
		return err
	}
	w.Write(data)
	return nil
}

// IssueSession stores the user's authentication state and profile in the
// session
func IssueSession() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		profile, err := user.ProfileFromContext(req.Context())
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
		j, err := json.Marshal(profile)
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
		var buf bytes.Buffer
		err = gzipWrite(&buf, []byte(j))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session.Values[sessionTokenKey] = string(buf.String())
		session.Save(w)
		http.Redirect(w, req, "/", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

// ContextFromSession populates the context with session information
func ContextFromSession(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := newContextFromSession(req.Context(), req)
		next.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func newContextFromSession(ctx context.Context, req *http.Request) context.Context {
	session, err := sessionStore.Get(req, sessionName)
	if err != nil {
		return ctx // No session
	}
	rawToken, ok := session.Values[sessionTokenKey].(string)
	var buf bytes.Buffer
	err = gunzipWrite(&buf, []byte(rawToken))
	if err != nil {
		log.Println(err)
		return ctx
	}
	if ok {
		token := oauth2.Token{}
		err = json.Unmarshal(buf.Bytes(), &token)
		if err != nil {
			log.Println(err)
		}
		ctx = context.WithValue(ctx, contextTokenKey, &token)
	}

	rawProfile, ok := session.Values[sessionProfileKey].(string)
	if ok {
		profile := user.Profile{}
		err = json.Unmarshal([]byte(rawProfile), &profile)
		if err != nil {
			log.Println(err)
		}
		ctx = user.WithProfile(ctx, &profile)
	}

	return ctx
}

// CallbackHandler handles Google redirection URI requests and adds the Google
// access token and Userinfoplus to the ctx. If authentication succeeds,
// handling delegates to the success handler, otherwise to the failure handler.
func CallbackHandler(config *oauth2.Config, fetcher user.Fetcher, success, failure http.Handler) http.Handler {
	success = oauth2Handler(config, fetcher, success, failure)
	return dgoauth2.CallbackHandler(config, success, failure)
}

// oauth2Handler is a http.Handler that gets the OAuth2 Token from the ctx
// to get the corresponding user profile. If successful, the user info
// is added to the ctx and the success handler is called. Otherwise, the
// failure handler is called.
func oauth2Handler(config *oauth2.Config, f user.Fetcher, success, failure http.Handler) http.Handler {
	if failure == nil {
		failure = gologin.DefaultFailureHandler
	}
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		token, err := dgoauth2.TokenFromContext(ctx)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		profile, err := f.Profile(ctx, config, token)
		err = validateResponse(profile, err)
		if err != nil {
			ctx = gologin.WithError(ctx, err)
			failure.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		ctx = user.WithProfile(ctx, profile)
		success.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// validateResponse returns an error if the given profile, raw
// http.Response, or error are unexpected. Returns nil if they are valid.
func validateResponse(profile *user.Profile, err error) error {
	if err != nil {
		return errors.Wrap(err, "unable to get Profile")
	}
	if profile == nil || profile.AccountName == "" {
		return errors.New("could not validate Profile")
	}
	return nil
}

// LogoutHandler logs a user out and deletes their session
func LogoutHandler(w http.ResponseWriter, req *http.Request) {
	destroySession(w)
	http.Redirect(w, req, "/", http.StatusFound)
}

func destroySession(w http.ResponseWriter) {
	sessionStore.Destroy(w, sessionName)
}
