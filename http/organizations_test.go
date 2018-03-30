package http

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	. "github.com/onsi/gomega"
	"github.com/pivotalservices/ignition/cloudfoundry/cloudfoundryfakes"
	"github.com/pivotalservices/ignition/internal"
	"github.com/pivotalservices/ignition/user"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"golang.org/x/oauth2"
)

func TestOrganizationHandler(t *testing.T) {
	spec.Run(t, "OrganizationHandler", testOrganizationHandler, spec.Report(report.Terminal{}))
}

func testOrganizationHandler(t *testing.T, when spec.G, it spec.S) {
	var s *httptest.Server
	var client *http.Client
	var req *http.Request
	var a *API
	var c *cloudfoundryfakes.FakeAPI

	it.Before(func() {
		RegisterTestingT(t)
		client = &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		c = &cloudfoundryfakes.FakeAPI{}
		token := &oauth2.Token{
			AccessToken:  "test-token",
			RefreshToken: "test-refresh-token",
			TokenType:    "bearer",
			Expiry:       time.Now().Add(time.Hour * 24),
		}
		s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.RequestURI, "/api") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				b := internal.HelperLoadBytes(t, "organization-response.json")
				w.Write(b)
			} else {
				profile := &user.Profile{
					AccountName: "testuser@test.com",
				}
				r = r.WithContext(WithToken(user.WithProfile(WithUserID(r.Context(), "test-userid"), profile), token))
				organizationHandler("http://example.net", "ignition", c).ServeHTTP(w, r)
			}
		}))
		a = &API{
			CCAPI: &cfclient.Client{
				Config: cfclient.Config{
					ApiAddress:   fmt.Sprintf("%s/api", s.URL),
					HttpClient:   http.DefaultClient,
					ClientID:     "cf",
					ClientSecret: "",
					Token:        "",
				},
			},
		}
		req, _ = http.NewRequest(http.MethodGet, s.URL, nil)
	})

	it.After(func() {
		s.Close()
	})

	when("the user has an org", func() {
		it("returns the org", func() {
			c.ListOrgsByQueryReturns([]cfclient.Org{
				cfclient.Org{
					Guid:                        "1234",
					Name:                        "test-org",
					CreatedAt:                   "now",
					UpdatedAt:                   "later",
					QuotaDefinitionGuid:         "321",
					DefaultIsolationSegmentGuid: "987",
				},
			}, nil)
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(resp.Body)
		})
	})

	when("the user has multiple orgs", func() {
		it("returns the org with the correct name", func() {
			c.ListOrgsByQueryReturns([]cfclient.Org{
				cfclient.Org{
					Guid:                        "1234",
					Name:                        "test-org",
					CreatedAt:                   "now",
					UpdatedAt:                   "later",
					QuotaDefinitionGuid:         "321",
					DefaultIsolationSegmentGuid: "987",
				},
			}, nil)
			resp, err := client.Do(req)
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})
}

func TestOrgName(t *testing.T) {
	spec.Run(t, "OrgName", testOrgName, spec.Report(report.Terminal{}))
}

func testOrgName(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	it("works with email addresses", func() {
		Expect(orgName("ignition", "test@example.net")).To(Equal("ignition-test"))
		Expect(orgName("igNiTion", "tEsT@example.net")).To(Equal("ignition-test"))
	})

	it("works with domain accounts", func() {
		Expect(orgName("ignition", "corp\\test")).To(Equal("ignition-test"))
		Expect(orgName("igNiTion", "corp\\tEsT")).To(Equal("ignition-test"))
	})

	it("works with plain accounts", func() {
		Expect(orgName("ignition", "test")).To(Equal("ignition-test"))
		Expect(orgName("igNiTion", "tEsT")).To(Equal("ignition-test"))
	})
}
