package openid_test

import (
	"context"
	"errors"
	"testing"
	"time"

	oidc "github.com/coreos/go-oidc"
	. "github.com/onsi/gomega"
	"github.com/pivotalservices/ignition/user/openid"
	"github.com/pivotalservices/ignition/user/openid/openidfakes"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"golang.org/x/oauth2"
)

func TestFetcher(t *testing.T) {
	spec.Run(t, "Fetcher", testFetcher, spec.Report(report.Terminal{}))
}

func testFetcher(t *testing.T, when spec.G, it spec.S) {
	var f *openid.Fetcher

	it.Before(func() {
		RegisterTestingT(t)
		f = &openid.Fetcher{}
	})

	when("using the OIDCIDVerifier", func() {
		var (
			o *openid.OIDCIDVerifier
			f *openidfakes.FakeOIDCVerifier
		)

		it.Before(func() {
			f = &openidfakes.FakeOIDCVerifier{}
			o = &openid.OIDCIDVerifier{
				Verifier: f,
			}
		})

		it("returns an error when the verifier fails", func() {
			f.VerifyReturns(nil, errors.New("test error"))
			c, err := o.Verify(context.Background(), "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("test error"))
			Expect(c).To(BeNil())
		})

		it("returns an error when claims are not set", func() {
			f.VerifyReturns(&oidc.IDToken{}, nil)
			c, err := o.Verify(context.Background(), "")
			Expect(err).To(HaveOccurred())
			Expect(c).To(BeNil())
		})
	})

	it("creates a valid verifier", func() {
		v := openid.NewVerifier("", "", "")
		Expect(v).NotTo(BeNil())
	})

	when("the verifier is nil", func() {
		it("returns an error", func() {
			p, err := f.Profile(context.Background(), nil, nil)
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("unable to verify token"))
			Expect(p).To(BeNil())
		})
	})

	when("the verifier is valid", func() {
		it.Before(func() {
			f.Verifier = openid.NewVerifier("", "", "")
		})

		when("the token is nil", func() {
			it("returns an error", func() {
				p, err := f.Profile(context.Background(), nil, nil)
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("unable to verify token"))
				Expect(p).To(BeNil())
			})
		})

		when("the token has no extra", func() {
			it("returns an error", func() {
				p, err := f.Profile(context.Background(), nil, &oauth2.Token{})
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("profile: no id_token"))
				Expect(p).To(BeNil())
			})
		})

		when("the token has includes id_token content", func() {
			var t *oauth2.Token

			it.Before(func() {
				t = &oauth2.Token{
					AccessToken:  "test-token",
					TokenType:    "bearer",
					RefreshToken: "test-refresh-token",
					Expiry:       time.Now().Add(24 * time.Hour),
				}
				extra := map[string]interface{}{
					"id_token": "test-id_token",
				}
				t = t.WithExtra(extra)
			})

			when("the token cannot be verified", func() {
				it("returns an error", func() {
					v := &openidfakes.FakeVerifier{}
					v.VerifyReturns(nil, errors.New("test error"))
					f.Verifier = v
					p, err := f.Profile(context.Background(), nil, t)
					Expect(err).NotTo(BeNil())
					Expect(err.Error()).To(ContainSubstring("unable to fetch claims: test error"))
					Expect(p).To(BeNil())
				})
			})

			when("the token can be verified", func() {
				it("returns the profile", func() {
					v := &openidfakes.FakeVerifier{}
					v.VerifyReturns(&openid.Claims{
						Email:      "test@example.net",
						UserName:   "tester",
						GivenName:  "Test",
						FamilyName: "User",
					}, nil)
					f.Verifier = v
					p, err := f.Profile(context.Background(), nil, t)
					Expect(err).To(BeNil())
					Expect(p).NotTo(BeNil())
					Expect(p.Name).To(Equal("Test User"))
					Expect(p.AccountName).To(Equal("tester"))
					Expect(p.Email).To(Equal("test@example.net"))
				})

				it("uses the email address as the account name if it is not set", func() {
					v := &openidfakes.FakeVerifier{}
					v.VerifyReturns(&openid.Claims{
						Email:      "test@example.net",
						UserName:   "",
						GivenName:  "Test",
						FamilyName: "User",
					}, nil)
					f.Verifier = v
					p, err := f.Profile(context.Background(), nil, t)
					Expect(err).To(BeNil())
					Expect(p).NotTo(BeNil())
					Expect(p.Name).To(Equal("Test User"))
					Expect(p.AccountName).To(Equal("test@example.net"))
					Expect(p.Email).To(Equal("test@example.net"))
				})
			})
		})
	})
}
