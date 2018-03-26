package cloudfoundry

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func helperLoadBytes(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func testInfo() *Info {
	return &Info{
		Name:                     "",
		Build:                    "",
		Support:                  "https://support.pivotal.io",
		Version:                  0,
		Description:              "",
		AuthorizationEndpoint:    "https://login.run.pcfbeta.io",
		TokenEndpoint:            "https://uaa.run.pcfbeta.io",
		MinCliVersion:            "6.23.0",
		MinRecommendedCliVersion: "6.23.0",
		APIVersion:               "2.103.0",
		AppSSHEndpoint:           "ssh.run.pcfbeta.io:2222",
		AppSSHHostKeyFingerprint: "00:f3:8b:eb:a8:d2:13:46:50:a1:02:49:d4:32:00:b8",
		AppSSHOauthClient:        "ssh-proxy",
		DopplerLoggingEndpoint:   "wss://doppler.run.pcfbeta.io:443",
		RoutingEndpoint:          "https://api.run.pcfbeta.io/routing",
	}
}

func TestInfo(t *testing.T) {
	spec.Run(t, "Info", func(t *testing.T, when spec.G, it spec.S) {
		var a *API

		it.Before(func() {
			RegisterTestingT(t)
		})

		when("the api URI is empty", func() {
			it.Before(func() {
				a = &API{}
			})

			it("returns an error", func() {
				api, err := a.Info()
				Expect(err).To(HaveOccurred())
				Expect(api).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("cannot get cloud controller info for empty URI"))
			})
		})

		when("there is an error making the info request", func() {
			it.Before(func() {
				a = &API{
					URI: "invalidurl",
				}
			})

			it("returns an error", func() {
				api, err := a.Info()
				Expect(err).To(HaveOccurred())
				Expect(api).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("could not get cloud controller info for URI:"))
			})
		})

		when("api info is returned", func() {
			var s *httptest.Server

			it.Before(func() {
				s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusOK)
					b := helperLoadBytes(t, "info-response.json")
					w.Write(b)
				}))
				a = &API{
					URI: s.URL,
				}
			})

			it.After(func() {
				s.Close()
			})

			it("returns the info", func() {
				api, err := a.Info()
				Expect(err).NotTo(HaveOccurred())
				Expect(api).NotTo(BeNil())
			})
		})

		when("invalid info is returned", func() {
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
				api, err := a.Info()
				Expect(err).To(HaveOccurred())
				Expect(api).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("could not decode response from URI:"))
			})
		})
	}, spec.Report(report.Terminal{}))
}
