package http

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestAPI(t *testing.T) {
	spec.Run(t, "API", testAPI, spec.Report(report.Terminal{}))
}

func testAPI(t *testing.T, when spec.G, it spec.S) {
	var api *API
	it.Before(func() {
		RegisterTestingT(t)
		api = &API{}
	})

	it("returns the URI correctly", func() {
		api.Scheme = "https"
		api.Domain = "example.net"
		Expect(api.URI()).To(Equal("https://example.net"))
		api.Port = 1234
		Expect(api.URI()).To(Equal("https://example.net:1234"))
	})

	it("creates a valid router", func() {
		r := api.createRouter()
		Expect(r).NotTo(BeNil())
		index := r.GetRoute("index")
		Expect(index).NotTo(BeNil())
		assets := r.GetRoute("assets")
		Expect(assets).NotTo(BeNil())
		nonexistent := r.GetRoute("nonexistent")
		Expect(nonexistent).To(BeNil())
	})
}
