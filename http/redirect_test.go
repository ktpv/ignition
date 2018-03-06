package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestEnsureHTTPS(t *testing.T) {
	spec.Run(t, "ensureHTTPS", func(t *testing.T, when spec.G, it spec.S) {
		var s *httptest.Server
		var client *http.Client
		var nextCalled bool
		var req *http.Request

		it.Before(func() {
			RegisterTestingT(t)
			client = &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse
				},
			}
			next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				nextCalled = true
			})
			s = httptest.NewServer(ensureHTTPS(next))
			req, _ = http.NewRequest(http.MethodGet, s.URL, nil)
		})

		it.After(func() {
			s.Close()
		})

		when("the user makes a request via http", func() {
			it("redirects the user to https", func() {
				resp, err := client.Do(req)
				Expect(nextCalled).To(BeFalse())
				Expect(err).To(BeNil())
				Expect(resp).NotTo(BeNil())
				Expect(resp.StatusCode).To(Equal(http.StatusMovedPermanently))
			})
		})

		when("the user makes a request via https", func() {
			it.Before(func() {
				req.Header.Set("X-Forwarded-Proto", "https")
			})

			it("calls the next handler", func() {
				resp, err := client.Do(req)
				Expect(nextCalled).To(BeTrue())
				Expect(err).To(BeNil())
				Expect(resp).NotTo(BeNil())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	}, spec.Report(report.Terminal{}))
}
