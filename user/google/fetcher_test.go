package google

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestFetcher(t *testing.T) {
	spec.Run(t, "Google Fetcher", func(t *testing.T, when spec.G, it spec.S) {
		it.Before(func() {
			RegisterTestingT(t)
		})
		when("fetching", func() {
			it("works", func() {
				Expect(true).To(BeTrue())
			})
		})
	}, spec.Report(report.Terminal{}))
}
