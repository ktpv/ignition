package cloudfoundry

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestOrganization(t *testing.T) {
	spec.Run(t, "Organization", func(t *testing.T, when spec.G, it spec.S) {
		var a *API

		it.Before(func() {
			RegisterTestingT(t)
		})

		when("the user has orgs", func() {
			var s *httptest.Server

			it.Before(func() {
				s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					b := helperLoadBytes(t, "organization-response.json")
					w.Write(b)
				}))
				a = &API{
					URI: s.URL,
				}
			})

			it.After(func() {
				s.Close()
			})

			it("should return the orgs", func() {
				api, err := a.Orgs()
				Expect(err).NotTo(HaveOccurred())
				Expect(api).NotTo(BeNil())
				fmt.Println((*api)[0].Guid)
			})
		})
	}, spec.Report(report.Terminal{}))
}
