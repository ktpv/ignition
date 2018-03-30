package http

import (
	"bytes"
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

func TestAuthorize(t *testing.T) {
	spec.Run(t, "Authorize", testAuthorize, spec.Report(report.Terminal{}))
}

func testAuthorize(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	it("is unauthorized when there is no token", func() {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		Authorize(nil, "").ServeHTTP(w, req)
		Expect(w.Code).To(Equal(http.StatusUnauthorized))
	})

	it("is unauthorized when there is an expired token", func() {
		w := httptest.NewRecorder()
		t := &oauth2.Token{
			AccessToken:  "test-token",
			RefreshToken: "test-refresh-token",
			TokenType:    "bearer",
			Expiry:       time.Now().Add(-24 * time.Hour),
		}
		ctx := context.WithValue(context.Background(), contextTokenKey, t)
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
		Authorize(nil, "").ServeHTTP(w, req)
		Expect(w.Code).To(Equal(http.StatusUnauthorized))
	})

	when("there is a valid token", func() {
		var (
			w   *httptest.ResponseRecorder
			req *http.Request
			t   *oauth2.Token
		)

		it.Before(func() {
			w = httptest.NewRecorder()
			t = &oauth2.Token{
				AccessToken:  "test-token",
				RefreshToken: "test-refresh-token",
				TokenType:    "bearer",
				Expiry:       time.Now().Add(24 * time.Hour),
			}
			ctx := WithToken(context.Background(), t)
			req = httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
		})

		it("is unauthorized if there is no profile in the context", func() {
			Authorize(nil, "").ServeHTTP(w, req)
			Expect(w.Code).To(Equal(http.StatusUnauthorized))
		})

		when("there is a valid profile in the context", func() {
			it.Before(func() {
				profile := &user.Profile{
					Name:        "Test User",
					Email:       "test@example.net",
					AccountName: "corp\tester",
				}
				req = req.WithContext(user.WithProfile(req.Context(), profile))
			})

			it("is forbidden if the user's email does not end with the domain", func() {
				Authorize(nil, "example.com").ServeHTTP(w, req)
				Expect(w.Code).To(Equal(http.StatusForbidden))
			})

			it("calls the next handler if the user's email does end with the domain", func() {
				called := false
				next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					called = true
					w.WriteHeader(http.StatusOK)
				})
				Authorize(next, "example.net").ServeHTTP(w, req)
				Expect(called).To(BeTrue())
				Expect(w.Code).To(Equal(http.StatusOK))
			})
		})
	})
}

func TestCompress(t *testing.T) {
	spec.Run(t, "Compress", testCompress, spec.Report(report.Terminal{}))
}

func testCompress(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	it("can round-trip a byte array", func() {
		b := bytes.NewBuffer(nil)
		gzipWrite(b, []byte("hello"))
		Expect(b.String()).NotTo(Equal("hello"))
		b2 := bytes.NewBuffer(nil)
		gunzipWrite(b2, b.Bytes())
		Expect(b2.String()).To(Equal("hello"))
	})
}

func TestIssueSession(t *testing.T) {
	spec.Run(t, "IssueSession", testIssueSession, spec.Report(report.Terminal{}))
}

func testIssueSession(t *testing.T, when spec.G, it spec.S) {
	var a *API
	it.Before(func() {
		RegisterTestingT(t)
		a = &API{
			SessionStore: newFakeSessionStore(),
		}
	})

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
}
