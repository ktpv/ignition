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

		it("does not validate when there is an empty client id", func() {
			err := api.validate("", "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("you must supply a non-empty client ID"))
			err = api.Run("", "")
			Expect(err).To(HaveOccurred())
		})

		it("does not validate when there is an empty client secret", func() {
			err := api.validate("abc", "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("you must supply a non-empty client secret"))
			err = api.Run("abc", "")
			Expect(err).To(HaveOccurred())
		})

		it("validates when there is a non empty client id and client secret", func() {
			err := api.validate("abc", "def")
			Expect(err).To(BeNil())
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
