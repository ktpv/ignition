package uaa_test

import (
	"context"
	"net/http"
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
			s = internal.TestDataServer(t, "empty-token.json", calledFunc)
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
			s = internal.TestDataServer(t, "token.json", calledFunc)
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

func TestUserIDForAccountName(t *testing.T) {
	spec.Run(t, "UserIDForAccountName", testUserIDForAccountName, spec.Report(report.Terminal{}))
}

func testUserIDForAccountName(t *testing.T, when spec.G, it spec.S) {
	var a *uaa.Client

	it.Before(func() {
		RegisterTestingT(t)
		a = &uaa.Client{}
	})

	it("cannot find a user id for an empty account name", func() {
		userID, err := a.UserIDForAccountName("")
		Expect(err).To(HaveOccurred())
		Expect(userID).To(BeZero())
		Expect(err.Error()).To(Equal("cannot search for a user with an empty account name"))
	})

	when("authentication is required", func() {
		var s *httptest.Server

		it.Before(func() {
			s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			}))
			a.URL = s.URL
		})

		it.After(func() {
			s.Close()
		})

		it("returns an error", func() {
			userID, err := a.UserIDForAccountName("test-user")
			Expect(err).To(HaveOccurred())
			Expect(userID).To(BeZero())
			Expect(err.Error()).To(ContainSubstring("uaa: cannot authenticate"))
		})
	})

	when("there is a valid token and a client", func() {
		var (
			s      *httptest.Server
			called bool
		)

		it.Before(func() {
			called = false

			a.Token = &oauth2.Token{
				AccessToken: "test-token",
				Expiry:      time.Now().Add(24 * time.Hour),
			}
			a.Client = http.DefaultClient
		})

		it.After(func() {
			s.Close()
		})

		when("a valid user is returned", func() {
			it.Before(func() {
				s = internal.TestDataServer(t, "users.json", func() {
					called = true
				})
				a.URL = s.URL
			})

			it("returns the user id", func() {
				userID, err := a.UserIDForAccountName("tester@pivotal.io")
				Expect(err).NotTo(HaveOccurred())
				Expect(userID).To(Equal("abcdef11-0000-dddd-aaaa-1234567890ab"))
				Expect(called).To(BeTrue())
			})
		})

		when("the users call fails", func() {
			it.Before(func() {
				s = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					called = true
					w.WriteHeader(http.StatusInternalServerError)
				}))
				a.URL = s.URL
			})

			it("returns the error", func() {
				userID, err := a.UserIDForAccountName("tester@pivotal.io")
				Expect(err).To(HaveOccurred())
				Expect(userID).To(BeZero())
				Expect(called).To(BeTrue())
			})
		})

		when("an empty user is returned", func() {
			it.Before(func() {
				s = internal.TestDataServer(t, "empty-user.json", func() {
					called = true
				})
				a.URL = s.URL
			})

			it("returns an error", func() {
				userID, err := a.UserIDForAccountName("tester@pivotal.io")
				Expect(err).To(HaveOccurred())
				Expect(userID).To(BeZero())
				Expect(err.Error()).To(Equal("cannot find user with account name: [tester@pivotal.io]"))
				Expect(called).To(BeTrue())
			})
		})
	})
}
