package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dghubble/gologin/google"
	dgoauth2 "github.com/dghubble/gologin/oauth2"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"golang.org/x/oauth2"
	oauth2v2 "google.golang.org/api/oauth2/v2"
)

func TestAPI(t *testing.T) {
	spec.Run(t, "API", func(t *testing.T, when spec.G, it spec.S) {
		var api *API
		it.Before(func() {
			RegisterTestingT(t)
			api = &API{}
		})

		it("does not validate when there is an empty client id", func() {
			err := api.validate("", "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("you must supply a non-empty client ID"))
			err = api.Run("", "")
			Expect(err).To(HaveOccurred())
		})

		it("does not validate when there is an empty client secret", func() {
			err := api.validate("abc", "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("you must supply a non-empty client secret"))
			err = api.Run("abc", "")
			Expect(err).To(HaveOccurred())
		})

		it("validates when there is a non empty client id and client secret", func() {
			err := api.validate("abc", "def")
			Expect(err).To(BeNil())
		})

		it("creates a valid router", func() {
			r := api.createRouter()
			Expect(r).NotTo(BeNil())
			index := r.GetRoute("index")
			Expect(index).NotTo(BeNil())
			assets := r.GetRoute("assets")
			Expect(assets).NotTo(BeNil())
			nonexistent := r.GetRoute("nonexistent")
			Expect(nonexistent).To(BeNil())
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
				ctx := google.WithUser(context.Background(), &oauth2v2.Userinfoplus{
					Email:      "test@pivotal.io",
					GivenName:  "Joe",
					FamilyName: "Tester",
				})
				req = req.WithContext(ctx)
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
				Expect(w.Code).Should(Equal(http.StatusInternalServerError))
			})

			it("issues a session", func() {
				handler := IssueSession()
				req := httptest.NewRequest("GET", "http://example.com/oauth2", nil)
				ctx := google.WithUser(context.Background(), &oauth2v2.Userinfoplus{
					Email:      "test@pivotal.io",
					GivenName:  "Joe",
					FamilyName: "Tester",
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
