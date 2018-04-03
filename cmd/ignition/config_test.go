package main

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

func resetEnv() {
	os.Unsetenv("IGNITION_AUTH_VARIANT")
	os.Unsetenv("IGNITION_CLIENT_ID")
	os.Unsetenv("IGNITION_CLIENT_SECRET")
	os.Unsetenv("IGNITION_AUTH_URL")
	os.Unsetenv("IGNITION_TOKEN_URL")
	os.Unsetenv("IGNITION_JWKS_URL")
	os.Unsetenv("IGNITION_ISSUER_URL")
	os.Unsetenv("IGNITION_AUTH_SCOPES")
	os.Unsetenv("IGNITION_AUTHORIZED_DOMAIN")
	os.Unsetenv("IGNITION_SESSION_SECRET")
	os.Unsetenv("IGNITION_PORT")
	os.Unsetenv("IGNITION_SERVE_PORT")
	os.Unsetenv("IGNITION_DOMAIN")
	os.Unsetenv("IGNITION_SCHEME")
	os.Unsetenv("IGNITION_WEB_ROOT")
	os.Unsetenv("IGNITION_UAA_URL")
	os.Unsetenv("IGNITION_APPS_URL")
	os.Unsetenv("IGNITION_CCAPI_URL")
	os.Unsetenv("IGNITION_CCAPI_CLIENT_ID")
	os.Unsetenv("IGNITION_CCAPI_CLIENT_SECRET")
	os.Unsetenv("IGNITION_CCAPI_USERNAME")
	os.Unsetenv("IGNITION_CCAPI_PASSWORD")
	os.Unsetenv("VCAP_APPLICATION")
	os.Unsetenv("VCAP_SERVICES")
	os.Unsetenv("PORT")
	os.Unsetenv("IGNITION_ORG_PREFIX")
	os.Unsetenv("IGNITION_QUOTA_ID")
	os.Unsetenv("IGNITION_UAA_ORIGIN")
}

func TestIgnitionMain(t *testing.T) {
	spec.Run(t, "NewAPI", func(t *testing.T, when spec.G, it spec.S) {
		currentDir, _ := os.Getwd()

		it.Before(func() {
			RegisterTestingT(t)
			resetEnv()
		})

		it.After(func() {
			resetEnv()
		})

		when("the environment is empty", func() {
			it("returns an error", func() {
				api, err := NewAPI()
				Expect(err).To(HaveOccurred())
				Expect(api).To(BeNil())
			})
		})

		when("all required env vars have been set", func() {
			it.Before(func() {
				os.Setenv("IGNITION_CLIENT_ID", "test-ignition-client-id")
				os.Setenv("IGNITION_CLIENT_SECRET", "test-ignition-client-secret")
				os.Setenv("IGNITION_AUTH_URL", "test-ignition-auth-url")
				os.Setenv("IGNITION_TOKEN_URL", "test-ignition-token-url")
				os.Setenv("IGNITION_JWKS_URL", "test-ignition-jwks-url")
				os.Setenv("IGNITION_ISSUER_URL", "test-ignition-issuer-url")
				os.Setenv("IGNITION_AUTHORIZED_DOMAIN", "test-ignition-authorized-domain")
				os.Setenv("IGNITION_SESSION_SECRET", "test-ignition-session-secret")
				os.Setenv("IGNITION_CCAPI_URL", "https://example.com")
				os.Setenv("IGNITION_UAA_URL", "https://example.com")
				os.Setenv("IGNITION_APPS_URL", "https://example.com")
				os.Setenv("IGNITION_CCAPI_USERNAME", "test-ccapi-username")
				os.Setenv("IGNITION_CCAPI_PASSWORD", "test-ccapi-password")
				os.Setenv("IGNITION_QUOTA_ID", "test-quotaid")
				os.Setenv("IGNITION_UAA_ORIGIN", "test-origin")
			})

			it("does not return an error", func() {
				api, err := NewAPI()
				Expect(err).NotTo(HaveOccurred())
				Expect(api).NotTo(BeNil())
			})

			it("fails if the ccapi url is empty", func() {
				os.Unsetenv("IGNITION_CCAPI_URL")
				api, err := NewAPI()
				Expect(err).To(HaveOccurred())
				Expect(api).To(BeNil())
			})

			it("fails if the client id is empty", func() {
				os.Unsetenv("IGNITION_CLIENT_ID")
				api, err := NewAPI()
				Expect(err).To(HaveOccurred())
				Expect(api).To(BeNil())
			})

			it("defaults the ccapi client id and secret to default values", func() {
				api, err := NewAPI()
				Expect(err).NotTo(HaveOccurred())
				Expect(api).NotTo(BeNil())
				Expect(api.APIConfig.ClientID).To(Equal("cf"))
				Expect(api.APIConfig.ClientSecret).To(Equal(""))
			})

			it("fails if the ccapi username is empty", func() {
				os.Unsetenv("IGNITION_CCAPI_USERNAME")
				api, err := NewAPI()
				Expect(err).To(HaveOccurred())
				Expect(api).To(BeNil())
			})

			it("fails if the client secret is empty", func() {
				os.Unsetenv("IGNITION_CLIENT_SECRET")
				api, err := NewAPI()
				Expect(err).To(HaveOccurred())
				Expect(api).To(BeNil())
			})

			it("fails if the ccapi password is empty", func() {
				os.Unsetenv("IGNITION_CCAPI_PASSWORD")
				api, err := NewAPI()
				Expect(err).To(HaveOccurred())
				Expect(api).To(BeNil())
			})

			it("fails if the quotaid is empty", func() {
				os.Unsetenv("IGNITION_QUOTA_ID")
				api, err := NewAPI()
				Expect(err).To(HaveOccurred())
				Expect(api).To(BeNil())
			})

			when("running on CF", func() {
				it.Before(func() {
					os.Setenv("PORT", "6543")
					os.Setenv("VCAP_APPLICATION", `{"cf_api": "https://api.run.pcfbeta.io","limits": {"fds": 16384},"application_name": "ignition","application_uris": ["ignition.pcfbeta.io"],"name": "ignition","space_name": "development","space_id": "test-space-id","uris": ["ignition.pcfbeta.io"],"users": null,"application_id": "test-app-id"}`)
					os.Setenv("VCAP_SERVICES", `{}`)
				})

				it("sets the webroot to the current directory", func() {
					api, err := NewAPI()
					Expect(err).NotTo(HaveOccurred())
					Expect(api).NotTo(BeNil())
					Expect(api.WebRoot).To(Equal(currentDir))
				})

				it("sets the port correctly", func() {
					api, err := NewAPI()
					Expect(err).NotTo(HaveOccurred())
					Expect(api).NotTo(BeNil())
					Expect(api.Port).To(Equal(443))
					Expect(api.ServePort).To(Equal(6543))
				})

				it("sets the scheme to https", func() {
					api, err := NewAPI()
					Expect(err).NotTo(HaveOccurred())
					Expect(api).NotTo(BeNil())
					Expect(api.Scheme).To(Equal("https"))
				})

				it("sets the domain", func() {
					api, err := NewAPI()
					Expect(err).NotTo(HaveOccurred())
					Expect(api).NotTo(BeNil())
					Expect(api.Domain).To(Equal("ignition.pcfbeta.io"))
				})

				it("returns an error if there are no application uris set", func() {
					os.Setenv("VCAP_APPLICATION", `{"cf_api": "https://api.run.pcfbeta.io","limits": {"fds": 16384},"application_name": "ignition","application_uris": [],"name": "ignition","space_name": "development","space_id": "test-space-id","uris": ["ignition.pcfbeta.io"],"users": null,"application_id": "test-app-id"}`)
					api, err := NewAPI()
					Expect(err).To(HaveOccurred())
					Expect(api).To(BeNil())
				})

				when("using the p-identity variant", func() {
					it.Before(func() {
						os.Setenv("IGNITION_AUTH_VARIANT", "p-identity")
					})

					it("fails if there is no service with the name identity", func() {
						api, err := NewAPI()
						Expect(err).To(HaveOccurred())
						Expect(api).To(BeNil())
						Expect(err.Error()).To(ContainSubstring("a Single Sign On service instance with the name \"identity\" is required to use this app"))
					})

					it("fails if client_id is not set", func() {
						os.Setenv("VCAP_SERVICES", `{
							"p-identity": [
								{
									"credentials": {
										"auth_domain": "https://ignition.login.run.pcfbeta.io",
										"client_secret": "test-cf-client-secret"
									},
									"syslog_drain_url": null,
									"volume_mounts": [],
									"label": "p-identity",
									"provider": null,
									"plan": "ignition",
									"name": "identity",
									"tags": []
								}
							]
						}`)

						api, err := NewAPI()
						Expect(err).To(HaveOccurred())
						Expect(api).To(BeNil())
						Expect(err.Error()).To(ContainSubstring("could not retrieve the client_id; make sure you have created and bound a Single Sign On service instance with the name \"identity\""))
					})

					it("fails if client_secret is not set", func() {
						os.Setenv("VCAP_SERVICES", `{
							"p-identity": [
								{
									"credentials": {
										"auth_domain": "https://ignition.login.run.pcfbeta.io",
										"client_id": "test-cf-client-id"
									},
									"syslog_drain_url": null,
									"volume_mounts": [],
									"label": "p-identity",
									"provider": null,
									"plan": "ignition",
									"name": "identity",
									"tags": []
								}
							]
						}`)

						api, err := NewAPI()
						Expect(err).To(HaveOccurred())
						Expect(api).To(BeNil())
						Expect(err.Error()).To(ContainSubstring("could not retrieve the client_secret; make sure you have created and bound a Single Sign On service instance with the name \"identity\""))
					})

					it("uses the client_id and client_secret values from the service binding", func() {
						os.Setenv("VCAP_SERVICES", `{
							"p-identity": [
								{
									"credentials": {
										"auth_domain": "https://ignition.login.run.pcfbeta.io",
										"client_id": "test-cf-client-id",
										"client_secret": "test-cf-client-secret"
									},
									"syslog_drain_url": null,
									"volume_mounts": [],
									"label": "p-identity",
									"provider": null,
									"plan": "ignition",
									"name": "identity",
									"tags": []
								}
							]
						}`)

						api, err := NewAPI()
						Expect(err).NotTo(HaveOccurred())
						Expect(api).NotTo(BeNil())
						Expect(api.UserConfig.ClientID).To(Equal("test-cf-client-id"))
						Expect(api.UserConfig.ClientSecret).To(Equal("test-cf-client-secret"))
					})
				})
			})

			when("not running on CF", func() {
				it("sets the webroot for local development", func() {
					api, err := NewAPI()
					Expect(err).NotTo(HaveOccurred())
					Expect(api).NotTo(BeNil())
					Expect(api.WebRoot).To(Equal(filepath.Join(currentDir, "web", "dist")))
				})

				it("uses the correct client id", func() {
					api, err := NewAPI()
					Expect(err).NotTo(HaveOccurred())
					Expect(api).NotTo(BeNil())
					Expect(api.UserConfig.ClientID).To(Equal("test-ignition-client-id"))
				})

				it("uses the correct client secret", func() {
					api, err := NewAPI()
					Expect(err).NotTo(HaveOccurred())
					Expect(api).NotTo(BeNil())
					Expect(api.UserConfig.ClientSecret).To(Equal("test-ignition-client-secret"))
				})
			})
		})
	}, spec.Report(report.Terminal{}))
}
