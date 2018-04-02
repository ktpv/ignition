package session

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	dgoauth2 "github.com/dghubble/gologin/oauth2"
	"github.com/dghubble/sessions"
	"github.com/pivotalservices/ignition/uaa"
	"github.com/pivotalservices/ignition/user"
	"golang.org/x/oauth2"
)

const (
	sessionTokenKey   = "token"
	sessionProfileKey = "profile"
	sessionEmailKey   = "email"
	sessionUAAIDKey   = "uaaid"
	sessionName       = "ignition"
)

// IssueSession stores the user's authentication state and profile in the
// session
func IssueSession(s sessions.Store, u uaa.API) http.Handler {
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

		session := s.New(sessionName)
		if session == nil {
			http.Error(w, "session cannot be created", http.StatusInternalServerError)
			return
		}
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
		err = GzipWrite(&buf, []byte(j))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		session.Values[sessionTokenKey] = string(buf.String())
		userID, err := u.UserIDForAccountName(profile.AccountName)
		if err == nil {
			session.Values[sessionUAAIDKey] = userID
		}
		session.Save(w)
		http.Redirect(w, req, "/", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}

// PopulateContext populates the context with session information
func PopulateContext(next http.Handler, s sessions.Store) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		session, err := s.Get(req, sessionName)
		if err != nil {
			next.ServeHTTP(w, req)
			return
		}
		rawToken, ok := session.Values[sessionTokenKey].(string)
		var buf bytes.Buffer
		err = GunzipWrite(&buf, []byte(rawToken))
		if err != nil {
			log.Println(err)
			next.ServeHTTP(w, req)
			return
		}
		ctx := req.Context()
		if ok {
			token := oauth2.Token{}
			err = json.Unmarshal(buf.Bytes(), &token)
			if err != nil {
				log.Println(err)
			}
			ctx = ContextWithToken(ctx, &token)
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
			ctx = ContextWithUserID(ctx, userID)
		}

		next.ServeHTTP(w, req.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// LogoutHandler logs a user out and deletes their session
func LogoutHandler(s sessions.Store) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		s.Destroy(w, sessionName)
		http.Redirect(w, req, "/", http.StatusFound)
	}
	return http.HandlerFunc(fn)
}
