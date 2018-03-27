package cloudfoundry

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"golang.org/x/oauth2"
)

func TestOrganization(t *testing.T) {
	spec.Run(t, "Organization", func(t *testing.T, when spec.G, it spec.S) {
		var a *API

		it.Before(func() {
			RegisterTestingT(t)
		})

		when("the user has orgs", func() {
			var s *httptest.Server
			var authorizationHeader string

			it.Before(func() {
				authorizationHeader = ""
				s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					authorizationHeader = r.Header.Get("Authorization")
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					b := helperLoadBytes(t, "organization-response.json")
					w.Write(b)
				}))
				a = &API{
					URI: s.URL,
					Config: &oauth2.Config{
						Endpoint: oauth2.Endpoint{
							AuthURL:  fmt.Sprintf("%s/oauth/authorize", s.URL),
							TokenURL: fmt.Sprintf("%s/oauth/token", s.URL),
						},
					},
					Token: &oauth2.Token{
						AccessToken:  "test-token",
						RefreshToken: "test-refresh-token",
						TokenType:    "bearer",
						Expiry:       time.Now().Add(time.Hour * 24),
					},
				}
			})

			it.After(func() {
				s.Close()
			})

			it("should return the orgs", func() {
				o, err := a.Organizations(context.Background())
				Expect(authorizationHeader).To(Equal("Bearer test-token"))
				Expect(err).NotTo(HaveOccurred())
				Expect(o).NotTo(BeNil())
				Expect(len(o)).To(Equal(1))
			})
		})
	}, spec.Report(report.Terminal{}))
}
