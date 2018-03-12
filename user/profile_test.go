package user

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func TestProfile(t *testing.T) {
	spec.Run(t, "profile", func(t *testing.T, when spec.G, it spec.S) {
		it.Before(func() {
			RegisterTestingT(t)
		})

		when("using an email address as the login name", func() {
			it("stores the profile in the context successfully", func() {
				expected := Profile{
					Email:       "test@example.net",
					AccountName: "test@example.net",
					Name:        "Test User",
				}
				ctx := WithProfile(context.Background(), &expected)
				Expect(ctx).NotTo(BeNil())
				Expect(ctx.Value(profileKey)).NotTo(BeNil())
				actual, err := ProfileFromContext(ctx)
				Expect(err).To(BeNil())
				Expect(actual).To(BeEquivalentTo(&expected))
				nonexistent, err := ProfileFromContext(context.Background())
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("Context missing Profile"))
				Expect(nonexistent).To(BeNil())
			})
		})
	})
}
