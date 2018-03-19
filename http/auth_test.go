package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dgoauth2 "github.com/dghubble/gologin/oauth2"
	"github.com/dghubble/sessions"
	. "github.com/onsi/gomega"
	"github.com/pivotalservices/ignition/user"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"golang.org/x/oauth2"
)

type fakeSessionStore struct {
	session       *sessions.Session
	newCalled     bool
	getCalled     bool
	saveCalled    bool
	destroyCalled bool
}

func newFakeSessionStore() *fakeSessionStore {
	return &fakeSessionStore{}
}

func (f *fakeSessionStore) New(name string) *sessions.Session {
	config := &sessions.Config{
		Path:     "/",
		MaxAge:   3600 * 24 * 7,
		HTTPOnly: true,
	}
	session := sessions.NewSession(f, name)
	session.Config = config
	f.session = session
	f.newCalled = true
	return f.session
}

func (f *fakeSessionStore) Get(req *http.Request, name string) (*sessions.Session, error) {
	f.getCalled = true
	return f.session, nil
}

func (f *fakeSessionStore) Save(w http.ResponseWriter, session *sessions.Session) error {
	f.saveCalled = true
	f.session = session
	return nil
}

func (f *fakeSessionStore) Destroy(w http.ResponseWriter, name string) {
	f.destroyCalled = true
	f.session = nil
}

func TestAuth(t *testing.T) {
	spec.Run(t, "auth", func(t *testing.T, when spec.G, it spec.S) {
		var a *API
		it.Before(func() {
			RegisterTestingT(t)
			a = &API{
				SessionStore: newFakeSessionStore(),
			}
		})
		when("issuing a session", func() {
			it("is an internal server error if the request has no user", func() {
				handler := a.IssueSession()
				req := httptest.NewRequest("GET", "http://example.com/oauth2", nil)
				ctx := context.Background()
				req = req.WithContext(ctx)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				Expect(w.Code).Should(Equal(http.StatusInternalServerError))
			})

			it("is an internal server error if the request has no token", func() {
				handler := a.IssueSession()
				req := httptest.NewRequest("GET", "http://example.com/oauth2", nil)
				ctx := user.WithProfile(context.Background(), &user.Profile{
					Email:       "test@pivotal.io",
					AccountName: "test@pivotal.io",
					Name:        "Joe Tester",
				})
				req = req.WithContext(ctx)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				Expect(w.Code).Should(Equal(http.StatusInternalServerError))
			})

			it("issues a session", func() {
				handler := a.IssueSession()
				req := httptest.NewRequest("GET", "http://example.com/oauth2", nil)
				ctx := user.WithProfile(context.Background(), &user.Profile{
					Email:       "test@pivotal.io",
					AccountName: "test@pivotal.io",
					Name:        "Joe Tester",
				})
				ctx = dgoauth2.WithToken(ctx, &oauth2.Token{
					TokenType:    "Bearer",
					AccessToken:  "1234",
					RefreshToken: "",
					Expiry:       time.Now().Add(3600 * time.Second),
				})
				req = req.WithContext(ctx)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				Expect(w.Code).Should(Equal(http.StatusFound))
			})
		})
	}, spec.Report(report.Terminal{}))
}
