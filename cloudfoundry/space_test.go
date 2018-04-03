package cloudfoundry_test

import (
	"errors"
	"testing"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	. "github.com/onsi/gomega"
	"github.com/pivotalservices/ignition/cloudfoundry"
	"github.com/pivotalservices/ignition/cloudfoundry/cloudfoundryfakes"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestCreateSpace(t *testing.T) {
	spec.Run(t, "CreateSpace", testCreateSpace, spec.Report(report.Terminal{}))
}

func testCreateSpace(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	it("returns an error if the creator returns an error", func() {
		a := &cloudfoundryfakes.FakeAPI{}
		a.CreateSpaceReturns(cfclient.Space{}, errors.New("test error"))
		err := cloudfoundry.CreateSpace("test-space", "test-organization-id", "test-user-id", a)
		Expect(err).To(HaveOccurred())
	})

	it("returns the space if it is created successfully", func() {
		a := &cloudfoundryfakes.FakeAPI{}
		a.CreateSpaceReturns(cfclient.Space{
			Guid:      "test-space-guid",
			Name:      "test-space",
			CreatedAt: "created-at",
			UpdatedAt: "updated-at",
		}, nil)
		err := cloudfoundry.CreateSpace("test-space", "test-organization-id", "test-user-id", a)
		Expect(err).NotTo(HaveOccurred())
	})
}
