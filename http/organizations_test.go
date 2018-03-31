package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	. "github.com/onsi/gomega"
	"github.com/pivotalservices/ignition/cloudfoundry/cloudfoundryfakes"
	"github.com/pivotalservices/ignition/user"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func TestOrganizationHandler(t *testing.T) {
	spec.Run(t, "OrganizationHandler", testOrganizationHandler, spec.Report(report.Terminal{}))
}

func testOrganizationHandler(t *testing.T, when spec.G, it spec.S) {
	var r *http.Request
	var w *httptest.ResponseRecorder
	var c *cloudfoundryfakes.FakeAPI

	it.Before(func() {
		RegisterTestingT(t)
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/", nil)
		c = &cloudfoundryfakes.FakeAPI{}
	})

	when("there is no profile in the context", func() {
		it("is not found", func() {
			r = httptest.NewRequest(http.MethodGet, "/", nil)
			organizationHandler("http://example.net", "ignition", "test-quota-id", c).ServeHTTP(w, r)
			Expect(w.Code).To(Equal(http.StatusNotFound))
		})
	})

	when("there is a profile in the context but no user id", func() {
		it("is not found", func() {
			r = httptest.NewRequest(http.MethodGet, "/", nil)
			profile := &user.Profile{
				AccountName: "testuser@test.com",
			}
			r = r.WithContext(user.WithProfile(r.Context(), profile))
			organizationHandler("http://example.net", "ignition", "test-quota-id", c).ServeHTTP(w, r)
			Expect(w.Code).To(Equal(http.StatusNotFound))
		})
	})

	when("there is a profile and a user id in the context", func() {
		it.Before(func() {
			r = httptest.NewRequest(http.MethodGet, "/", nil)
			profile := &user.Profile{
				AccountName: "testuser@test.com",
			}
			r = r.WithContext(user.WithProfile(WithUserID(r.Context(), "test-user-id"), profile))
		})

		when("orgs cannot be retrieved", func() {
			it.Before(func() {
				c.ListOrgsByQueryReturns(nil, errors.New("test error"))
			})

			it("is an internal server error", func() {
				organizationHandler("http://example.net", "ignition", "test-quota-id", c).ServeHTTP(w, r)
				Expect(w.Code).To(Equal(http.StatusInternalServerError))
			})
		})

		when("there are no orgs for the user", func() {
			it.Before(func() {
				c.ListOrgsByQueryReturns(nil, nil)
			})

			it("is not found", func() {
				organizationHandler("http://example.net", "ignition", "test-quota-id", c).ServeHTTP(w, r)
				Expect(w.Code).To(Equal(http.StatusNotFound))
			})
		})

		when("there are multiple orgs for the user", func() {
			it.Before(func() {
				c.ListOrgsByQueryReturns([]cfclient.Org{
					cfclient.Org{
						Guid:                        "test-org-2",
						Name:                        "ignition-testuser1",
						QuotaDefinitionGuid:         "ignition-quota2-id",
						DefaultIsolationSegmentGuid: "default-iso-guid",
						CreatedAt:                   "created-at",
						UpdatedAt:                   "updated-at",
					},
					cfclient.Org{
						Guid:                        "test-org-1",
						Name:                        "ignition-testuser",
						QuotaDefinitionGuid:         "ignition-quota-id",
						DefaultIsolationSegmentGuid: "default-iso-guid",
						CreatedAt:                   "created-at",
						UpdatedAt:                   "updated-at",
					},
				}, nil)
			})

			it("selects the correct org when there is a name match", func() {
				organizationHandler("http://example.net", "ignition", "test-quota2-id", c).ServeHTTP(w, r)
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("test-org-1"))
			})

			it("is not found when there is no name or quota match", func() {
				organizationHandler("http://example.net", "ignition1", "test-quota2-id", c).ServeHTTP(w, r)
				Expect(w.Code).To(Equal(http.StatusNotFound))
			})

			it("selects the correct org when there is a quota match (but not a name match)", func() {
				organizationHandler("http://example.net", "ignition2", "ignition-quota-id", c).ServeHTTP(w, r)
				Expect(w.Code).To(Equal(http.StatusOK))
				Expect(w.Body.String()).To(ContainSubstring("test-org-1"))
			})
		})
	})
}

func TestOrgName(t *testing.T) {
	spec.Run(t, "OrgName", testOrgName, spec.Report(report.Terminal{}))
}

func testOrgName(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	it("works with email addresses", func() {
		Expect(orgName("ignition", "test@example.net")).To(Equal("ignition-test"))
		Expect(orgName("igNiTion", "tEsT@example.net")).To(Equal("ignition-test"))
	})

	it("works with domain accounts", func() {
		Expect(orgName("ignition", "corp\\test")).To(Equal("ignition-test"))
		Expect(orgName("igNiTion", "corp\\tEsT")).To(Equal("ignition-test"))
	})

	it("works with plain accounts", func() {
		Expect(orgName("ignition", "test")).To(Equal("ignition-test"))
		Expect(orgName("igNiTion", "tEsT")).To(Equal("ignition-test"))
	})
}
