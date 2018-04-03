package session_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dgoauth2 "github.com/dghubble/gologin/oauth2"
	"github.com/dghubble/sessions"
	. "github.com/onsi/gomega"
	"github.com/pivotalservices/ignition/http/session"
	"github.com/pivotalservices/ignition/http/session/sessionfakes"
	"github.com/pivotalservices/ignition/uaa/uaafakes"
	"github.com/pivotalservices/ignition/user"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"golang.org/x/oauth2"
)

func TestIssueSession(t *testing.T) {
	spec.Run(t, "IssueSession", testIssueSession, spec.Report(report.Terminal{}))
}

func testIssueSession(t *testing.T, when spec.G, it spec.S) {
	var (
		fakeUAAAPI       *uaafakes.FakeAPI
		fakeSessionStore *sessionfakes.FakeStore
		s                *sessions.Session
	)

	it.Before(func() {
		RegisterTestingT(t)
		fakeUAAAPI = &uaafakes.FakeAPI{}
		fakeSessionStore = &sessionfakes.FakeStore{}
		s = sessions.NewSession(fakeSessionStore, "ignition-test")
		fakeSessionStore.NewReturns(s)
		fakeSessionStore.SaveReturns(nil)
		fakeSessionStore.GetReturns(s, nil)
	})

	when("there is no user profile", func() {
		it("is an internal server error", func() {
			handler := session.IssueSession(fakeSessionStore, fakeUAAAPI)
			req := httptest.NewRequest("GET", "http://example.com/oauth2", nil)
			ctx := context.Background()
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			Expect(w.Code).Should(Equal(http.StatusInternalServerError))
		})
	})

	when("there is no token", func() {
		it("is an internal server error", func() {
			handler := session.IssueSession(fakeSessionStore, fakeUAAAPI)
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
	})

	when("there is a user profile and a token", func() {
		var ctx context.Context

		it.Before(func() {
			ctx = user.WithProfile(context.Background(), &user.Profile{
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
		})

		it("is an internal server error if the session cannot be created", func() {
			fakeSessionStore.NewReturns(nil)
			handler := session.IssueSession(fakeSessionStore, fakeUAAAPI)
			req := httptest.NewRequest("GET", "http://example.com/oauth2", nil).WithContext(ctx)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			Expect(w.Code).Should(Equal(http.StatusInternalServerError))
		})

		it("issues a session", func() {
			handler := session.IssueSession(fakeSessionStore, fakeUAAAPI)
			req := httptest.NewRequest("GET", "http://example.com/oauth2", nil).WithContext(ctx)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			Expect(w.Code).Should(Equal(http.StatusFound))
		})

		when("there is no user ID for the account name", func() {
			it.Before(func() {
				fakeUAAAPI.UserIDForAccountNameReturns("", errors.New("test error"))
			})

			it("does not store the user ID in the session", func() {
				handler := session.IssueSession(fakeSessionStore, fakeUAAAPI)
				req := httptest.NewRequest("GET", "http://example.com/oauth2", nil).WithContext(ctx)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				Expect(w.Code).Should(Equal(http.StatusFound))
				Expect(s.Values).NotTo(HaveKeyWithValue("uaaid", "test-user-id"))
			})
		})

		when("there is a user ID for the account name", func() {
			it.Before(func() {
				fakeUAAAPI.UserIDForAccountNameReturns("test-user-id", nil)
			})

			it("stores the user ID in the session", func() {
				handler := session.IssueSession(fakeSessionStore, fakeUAAAPI)
				req := httptest.NewRequest("GET", "http://example.com/oauth2", nil).WithContext(ctx)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				Expect(w.Code).Should(Equal(http.StatusFound))
				Expect(s.Values).To(HaveKeyWithValue("uaaid", "test-user-id"))
			})
		})
	})
}

func TestPopulateContext(t *testing.T) {
	spec.Run(t, "PopulateContext", testPopulateContext, spec.Report(report.Terminal{}))
}

func testPopulateContext(t *testing.T, when spec.G, it spec.S) {
	var (
		fakeSessionStore *sessionfakes.FakeStore
		s                *sessions.Session
	)

	it.Before(func() {
		RegisterTestingT(t)
		fakeSessionStore = &sessionfakes.FakeStore{}
		s = sessions.NewSession(fakeSessionStore, "ignition-test")

		fakeSessionStore.NewReturns(s)
		fakeSessionStore.SaveReturns(nil)
		fakeSessionStore.GetReturns(s, nil)
	})

	it("invokes the next handler", func() {
		called := false
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true })
		handler := session.PopulateContext(next, fakeSessionStore)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		handler.ServeHTTP(w, r)
		Expect(called).To(BeTrue())
	})

	when("there is a token, profile, and user ID", func() {
		var (
			nextContext context.Context
			handler     http.Handler
		)

		it.Before(func() {
			b := bytes.NewBuffer(nil)
			session.GzipWrite(b, []byte(`{"access_token": "test-token"}`))
			s.Values["token"] = b.String()
			s.Values["profile"] = `{"email": "test@pivotal.io"}`
			s.Values["uaaid"] = "testuser"

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { nextContext = r.Context() })
			handler = session.PopulateContext(next, fakeSessionStore)
		})

		when("the session cannot be retrieved", func() {
			it.Before(func() {
				fakeSessionStore.GetReturns(nil, errors.New("test error"))
				w := httptest.NewRecorder()
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				handler.ServeHTTP(w, r)
			})

			it("doesn't add anything to the context", func() {
				_, err := session.TokenFromContext(nextContext)
				Expect(err).To(HaveOccurred())
				_, err = session.UserIDFromContext(nextContext)
				Expect(err).To(HaveOccurred())
				_, err = user.ProfileFromContext(nextContext)
				Expect(err).To(HaveOccurred())
			})
		})

		when("the session can be retrieved", func() {
			it.Before(func() {
				w := httptest.NewRecorder()
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				handler.ServeHTTP(w, r)
			})

			it("adds the token to the request context", func() {
				t, err := session.TokenFromContext(nextContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(t.AccessToken).To(Equal("test-token"))
			})

			it("adds the profile to the request context", func() {
				profile, err := user.ProfileFromContext(nextContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(profile.Email).To(Equal("test@pivotal.io"))
			})

			it("adds the user ID to the request context", func() {
				userID, err := session.UserIDFromContext(nextContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(userID).To(Equal("testuser"))
			})
		})
	})
}

func TestLogoutHandler(t *testing.T) {
	spec.Run(t, "LogoutHandler", testLogoutHandler, spec.Report(report.Terminal{}))
}

func testLogoutHandler(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	it("destroys the session", func() {
		w := httptest.NewRecorder()
		s := &sessionfakes.FakeStore{}
		handler := session.LogoutHandler(s)
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
		Expect(s.DestroyCallCount()).To(Equal(1))
		Expect(w.Code).To(Equal(http.StatusFound))
		Expect(w.Header().Get("Location")).To(Equal("/"))
	})
}

func TestUpdateSessionWithUserID(t *testing.T) {
	spec.Run(t, "UpdateSessionWithUserID", testUpdateSessionWithUserID, spec.Report(report.Terminal{}))
}

func testUpdateSessionWithUserID(t *testing.T, when spec.G, it spec.S) {
	var (
		fakeSessionStore *sessionfakes.FakeStore
		s                *sessions.Session
	)

	it.Before(func() {
		RegisterTestingT(t)
		fakeSessionStore = &sessionfakes.FakeStore{}
		s = sessions.NewSession(fakeSessionStore, "ignition-test")

		fakeSessionStore.NewReturns(s)
		fakeSessionStore.SaveReturns(nil)
		fakeSessionStore.GetReturns(s, nil)
	})

	it("saves the user id when it is non-zero", func() {
		w := httptest.NewRecorder()
		session.UpdateSessionWithUserID(w, httptest.NewRequest(http.MethodGet, "/", nil), fakeSessionStore, "test-user-id")
		Expect(fakeSessionStore.SaveCallCount()).To(Equal(1))
		Expect(s.Values).To(HaveKeyWithValue("uaaid", "test-user-id"))
	})

	it("does not save an empty user id", func() {
		w := httptest.NewRecorder()
		session.UpdateSessionWithUserID(w, httptest.NewRequest(http.MethodGet, "/", nil), fakeSessionStore, "")
		Expect(fakeSessionStore.SaveCallCount()).To(Equal(0))
		Expect(s.Values).NotTo(HaveKeyWithValue("uaaid", ""))
	})

	it("does not save if the session cannot be retrieved", func() {
		w := httptest.NewRecorder()
		fakeSessionStore.GetReturns(nil, errors.New("test error"))
		session.UpdateSessionWithUserID(w, httptest.NewRequest(http.MethodGet, "/", nil), fakeSessionStore, "test-user-id")
		Expect(fakeSessionStore.SaveCallCount()).To(Equal(0))
		Expect(s.Values).NotTo(HaveKeyWithValue("uaaid", ""))
	})
}
