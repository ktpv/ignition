package uaa_test

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/pivotalservices/ignition/internal"
	"github.com/pivotalservices/ignition/uaa"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"golang.org/x/oauth2"
)

func TestAuthenticate(t *testing.T) {
	spec.Run(t, "Authenticate", testAuthenticate, spec.Report(report.Terminal{}))
}

func testAuthenticate(t *testing.T, when spec.G, it spec.S) {
	var a *uaa.Client

	it.Before(func() {
		RegisterTestingT(t)
	})

	when("there is a valid client and token", func() {
		it.Before(func() {
			config := &oauth2.Config{
				ClientID:     "cf",
				ClientSecret: "",
			}
			t := &oauth2.Token{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
				Expiry:       time.Now().Add(24 * time.Hour),
			}
			a = &uaa.Client{
				Client: config.Client(context.Background(), t),
				Token:  t,
			}
		})

		it("succeeds", func() {
			err := a.Authenticate()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	when("there is a need to authenticate but the token is empty", func() {
		var (
			s      *httptest.Server
			called bool
		)

		it.Before(func() {
			called = false
			calledFunc := func() {
				called = true
			}
			s = internal.ServeFromTestdata(t, "empty-token.json", calledFunc)
			a = &uaa.Client{
				URL:          s.URL,
				ClientID:     "cf",
				ClientSecret: "",
				Username:     "test-user",
				Password:     "test-password",
			}
		})

		it.After(func() {
			s.Close()
		})

		it("returns an error", func() {
			err := a.Authenticate()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not retrieve UAA token"))
			Expect(called).To(BeTrue())
		})
	})

	when("there is a need to authenticate and the token is valid", func() {
		var (
			s      *httptest.Server
			called bool
		)

		it.Before(func() {
			called = false
			calledFunc := func() {
				called = true
			}
			s = internal.ServeFromTestdata(t, "token.json", calledFunc)
			a = &uaa.Client{
				URL:          s.URL,
				ClientID:     "cf",
				ClientSecret: "",
				Username:     "test-user",
				Password:     "test-password",
			}
		})

		it.After(func() {
			s.Close()
		})

		it("succeeds", func() {
			err := a.Authenticate()
			Expect(err).NotTo(HaveOccurred())
			Expect(called).To(BeTrue())
		})
	})
}
