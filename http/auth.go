package http

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/uaa-cli/uaa"
	"github.com/dghubble/gologin"
	dgoauth2 "github.com/dghubble/gologin/oauth2"
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
	sessionUAAIDKey       = "uaaid"
	sessionName           = "ignition"
	contextTokenKey   key = iota
	contextUserIDKey  key = iota
)

func (a *API) handleAuth(r *mux.Router) {
	stateConfig := gologin.DefaultCookieConfig
	if a.Domain == "localhost" {
		stateConfig = gologin.DebugOnlyCookieConfig
	}
	r.Handle("/login", ensureHTTPS(dgoauth2.StateHandler(stateConfig, dgoauth2.LoginHandler(a.UserConfig, nil)))).Name("login")
	r.Handle("/oauth2", ensureHTTPS(dgoauth2.StateHandler(stateConfig, CallbackHandler(a.UserConfig, a.Fetcher, a.IssueSession(), http.HandlerFunc(a.LogoutHandler))))).Name("oauth2")
	r.Handle("/logout", ensureHTTPS(http.HandlerFunc(a.LogoutHandler))).Name("logout")
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

// gzipWrite reads from the slice of bytes and writes the compressed data to the
// writer
func gzipWrite(w io.Writer, data []byte) error {
	// Write gzipped data to the client
	gw, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		return err
	}
	defer gw.Close()
	gw.Write(data)
	return err
}

// gunzipWrite reads from the gzipped slice of bytes and writes the uncompressed
// data to the writer
func gunzipWrite(w io.Writer, data []byte) error {
	// Write gzipped data to the client
	gr, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return err
	}
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
func (a *API) IssueSession() http.Handler {
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

		session := a.SessionStore.New(sessionName)
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
		userid, ok, _ := a.SearchForUserWithAccountName(req.Context(), profile.AccountName)
		if ok {
			session.Values[sessionUAAIDKey] = userid
		}
		session.Save(w)
		http.Redirect(w, req, "/", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

// SearchForUserWithAccountName queries the UAA API for users filtered by account name
func (a *API) SearchForUserWithAccountName(ctx context.Context, accountName string) (string, bool, error) {
	if strings.TrimSpace(accountName) == "" {
		return "", false, errors.New("cannot search for a user with an empty account name")
	}
	oauthConfig := oauth2.Config{
		ClientID:     "cf",
		ClientSecret: "",
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth/authorize", a.UAAURL),
			TokenURL: fmt.Sprintf("%s/oauth/token", a.UAAURL),
		},
	}

	t, err := oauthConfig.PasswordCredentialsToken(ctx, a.APIUsername, a.APIPassword)
	if err != nil {
		return "", false, errors.Wrap(err, "could not retrieve UAA token")
	}
	client := oauthConfig.Client(ctx, t)

	uaactx := uaa.NewContextWithToken(t.AccessToken)
	uaaConfig := uaa.NewConfig()
	uaaConfig.AddTarget(uaa.Target{
		BaseUrl: a.UAAURL,
	})
	uaaConfig.AddContext(uaactx)
	userManager := uaa.UserManager{
		Config:     uaaConfig,
		HttpClient: client,
	}
	u, err := userManager.GetByUsername(accountName, "", "")
	if err != nil || strings.TrimSpace(u.ID) == "" {
		return "", false, err
	}
	return u.ID, true, nil
}

// ContextFromSession populates the context with session information
func (a *API) ContextFromSession(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		ctx := a.newContextFromSession(req.Context(), req)
		next.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func (a *API) newContextFromSession(ctx context.Context, req *http.Request) context.Context {
	session, err := a.SessionStore.Get(req, sessionName)
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
		ctx = WithToken(ctx, &token)
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
	userID, ok := session.Values[sessionUAAIDKey].(string)
	if ok {
		ctx = WithUserID(ctx, userID)
	}

	return ctx
}

// TokenFromContext returns the Token from the ctx.
func TokenFromContext(ctx context.Context) (*oauth2.Token, error) {
	token, ok := ctx.Value(contextTokenKey).(*oauth2.Token)
	if !ok {
		return nil, fmt.Errorf("context missing Token")
	}
	return token, nil
}

// WithToken returns a copy of ctx that stores the Token.
func WithToken(ctx context.Context, token *oauth2.Token) context.Context {
	return context.WithValue(ctx, contextTokenKey, token)
}

// UserIDFromContext returns the user ID from the ctx.
func UserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(contextUserIDKey).(string)
	if !ok {
		return "", fmt.Errorf("context missing UserID")
	}
	return userID, nil
}

// WithUserID returns a copy of ctx that stores the user ID.
func WithUserID(ctx context.Context, userID string) context.Context {
	if strings.TrimSpace(userID) == "" {
		return ctx
	}
	return context.WithValue(ctx, contextUserIDKey, userID)
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
func (a *API) LogoutHandler(w http.ResponseWriter, req *http.Request) {
	a.destroySession(w)
	http.Redirect(w, req, "/", http.StatusFound)
}

func (a *API) destroySession(w http.ResponseWriter) {
	a.SessionStore.Destroy(w, sessionName)
}
