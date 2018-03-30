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

func TestOrg(t *testing.T) {
	spec.Run(t, "Org", testOrg, spec.Report(report.Terminal{}))
}

func testOrg(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	it("returns an error if the querier returns an error", func() {
		a := &cloudfoundryfakes.FakeAPI{}
		a.ListOrgsByQueryReturns(nil, errors.New("test error"))
		orgs, err := cloudfoundry.OrgsForUserID("123", "", a)
		Expect(err).To(HaveOccurred())
		Expect(orgs).To(BeNil())
	})

	it("maps the cfclient.Org to the Organization correctly", func() {
		a := &cloudfoundryfakes.FakeAPI{}
		a.ListOrgsByQueryReturns([]cfclient.Org{
			cfclient.Org{
				Guid:                        "1234",
				Name:                        "test-org",
				CreatedAt:                   "now",
				UpdatedAt:                   "later",
				QuotaDefinitionGuid:         "321",
				DefaultIsolationSegmentGuid: "987",
			},
		}, nil)
		orgs, err := cloudfoundry.OrgsForUserID("123", "https://example.com", a)
		Expect(err).NotTo(HaveOccurred())
		Expect(orgs).NotTo(BeNil())
		Expect(len(orgs)).To(Equal(1))
		Expect(orgs[0].GUID).To(Equal("1234"))
		Expect(orgs[0].CreatedAt).To(Equal("now"))
		Expect(orgs[0].UpdatedAt).To(Equal("later"))
		Expect(orgs[0].Name).To(Equal("test-org"))
		Expect(orgs[0].QuotaDefinitionGUID).To(Equal("321"))
		Expect(orgs[0].DefaultIsolationSegmentGUID).To(Equal("987"))
		Expect(orgs[0].URL).To(Equal("https://example.com/organizations/1234"))
	})
}
