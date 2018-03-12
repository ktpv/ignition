package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	dgoauth2 "github.com/dghubble/gologin/oauth2"
	. "github.com/onsi/gomega"
	"github.com/pivotalservices/ignition/user"
	"github.com/sclevine/spec"
	"golang.org/x/oauth2"
)

func TestAuth(t *testing.T) {
	spec.Run(t, "auth", func(t *testing.T, when spec.G, it spec.S) {
		it.Before(func() {
			RegisterTestingT(t)
		})
		when("issuing a session", func() {
			it("is an internal server error if the request has no user", func() {
				handler := IssueSession()
				req := httptest.NewRequest("GET", "http://example.com/oauth2", nil)
				ctx := context.Background()
				req = req.WithContext(ctx)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				Expect(w.Code).Should(Equal(http.StatusInternalServerError))
			})

			it("is an internal server error if the request has no token", func() {
				handler := IssueSession()
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
				handler := IssueSession()
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
	})
}
