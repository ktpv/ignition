package cloudfoundry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestAPI(t *testing.T) {
	spec.Run(t, "API", func(t *testing.T, when spec.G, it spec.S) {
		var a *API

		it.Before(func() {
			RegisterTestingT(t)
		})

		when("info is not returned from the server", func() {
			var s *httptest.Server

			it.Before(func() {
				s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
				}))
				a = &API{
					URI: s.URL,
				}
			})

			it.After(func() {
				s.Close()
			})

			it("returns an error", func() {
				err := a.Authenticate()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not decode response from URI:"))
			})
		})

		when("info is returned from the server", func() {
			var s *httptest.Server
			var m map[string]string

			it.Before(func() {
				m = make(map[string]string, 0)
				m["/v2/info"] = "info-response.json"
				m["/login"] = "login-response.json"
				m["/login/oauth/token"] = "token-response.json"
				s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					if r.URL.String() == "/v2/info" {
						i := testInfo()
						i.AuthorizationEndpoint = fmt.Sprintf("%s/login", s.URL)
						json.NewEncoder(w).Encode(i)
					}

					if r.URL.String() == "/login" {
						l := &login{
							Links: struct {
								UAA   string `json:"uaa"`
								Login string `json:"login"`
							}{
								Login: fmt.Sprintf("%s/login", s.URL),
							},
						}
						json.NewEncoder(w).Encode(l)
					}

					file := m[r.URL.String()]
					b := helperLoadBytes(t, file)
					w.Write(b)
				}))
				a = &API{
					URI: s.URL,
				}
			})

			it.After(func() {
				s.Close()
			})

			it("does not an error", func() {
				err := a.Authenticate()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	}, spec.Report(report.Terminal{}))
}
