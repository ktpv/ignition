package session_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/pivotalservices/ignition/http/session"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
	"golang.org/x/oauth2"
)

func TestContext(t *testing.T) {
	spec.Run(t, "Context", testContext, spec.Report(report.Terminal{}))
}

func testContext(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	it("errors when there is no token", func() {
		token, err := session.TokenFromContext(context.Background())
		Expect(err).To(HaveOccurred())
		Expect(token).To(BeNil())
		Expect(err.Error()).To(Equal("context missing Token"))
	})

	it("round trips a token", func() {
		ctx := session.ContextWithToken(context.Background(), &oauth2.Token{
			AccessToken: "test-token",
		})
		token, err := session.TokenFromContext(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(token).NotTo(BeNil())
		Expect(token.AccessToken).To(Equal("test-token"))
	})
}

func TestUserID(t *testing.T) {
	spec.Run(t, "UserID", testUserID, spec.Report(report.Terminal{}))
}

func testUserID(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	it("errors when there is no user ID", func() {
		userID, err := session.UserIDFromContext(context.Background())
		Expect(err).To(HaveOccurred())
		Expect(userID).To(BeZero())
		Expect(err.Error()).To(Equal("context missing UserID"))
	})

	it("round trips a user ID", func() {
		ctx := session.ContextWithUserID(context.Background(), "test-user-id")
		userID, err := session.UserIDFromContext(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(userID).To(Equal("test-user-id"))
	})

	it("errors when you round trip after storing an empty user ID", func() {
		ctx := session.ContextWithUserID(context.Background(), "")
		userID, err := session.UserIDFromContext(ctx)
		Expect(err).To(HaveOccurred())
		Expect(userID).To(BeZero())
		Expect(err.Error()).To(Equal("context missing UserID"))
	})
}
