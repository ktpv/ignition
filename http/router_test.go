package http

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func TestAPI(t *testing.T) {
	spec.Run(t, "API", func(t *testing.T, when spec.G, it spec.S) {
		var api *API
		it.Before(func() {
			RegisterTestingT(t)
			api = &API{}
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
	})
}
