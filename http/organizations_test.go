package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	. "github.com/onsi/gomega"
	"github.com/pivotalservices/ignition/user"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"golang.org/x/oauth2"
)

func helperLoadBytes(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

type fakeQuerier struct{}

func (q *fakeQuerier) ListOrgsByQuery(query url.Values) ([]cfclient.Org, error) {
	return []cfclient.Org{
		cfclient.Org{
			Guid: "1234",
			Name: "test-org",
		},
	}, nil
}

func TestOrganizationHandler(t *testing.T) {
	spec.Run(t, "OrganizationHandler", func(t *testing.T, when spec.G, it spec.S) {
		var s *httptest.Server
		var client *http.Client
		var req *http.Request
		var a *API

		it.Before(func() {
			RegisterTestingT(t)
			client = &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
			token := &oauth2.Token{
				AccessToken:  "test-token",
				RefreshToken: "test-refresh-token",
				TokenType:    "bearer",
				Expiry:       time.Now().Add(time.Hour * 24),
			}
			s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Println(r.RequestURI)
				if strings.HasPrefix(r.RequestURI, "/api") {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					b := helperLoadBytes(t, "organization-response.json")
					w.Write(b)
				} else {
					profile := &user.Profile{
						AccountName: "testuser@test.com",
					}
					r = r.WithContext(WithToken(user.WithProfile(WithUserID(r.Context(), "test-userid"), profile), token))
					organizationHandler("http://example.net", &fakeQuerier{}).ServeHTTP(w, r)
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
				resp, err := client.Do(req)
				Expect(err).To(BeNil())
				Expect(resp).NotTo(BeNil())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	}, spec.Report(report.Terminal{}))
}
