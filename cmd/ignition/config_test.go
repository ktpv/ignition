package main

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func TestAPI(t *testing.T) {
	spec.Run(t, "Config", func(t *testing.T, when spec.G, it spec.S) {
		var (
			authVariant     string
			clientID        string
			clientSecret    string
			issuerURL       string
			authURL         string
			tokenURL        string
			jwksURL         string
			expectedWebRoot string
		)

		it.Before(func() {
			RegisterTestingT(t)
			authVariant = os.Getenv("IGNITION_AUTH_VARIANT")
			clientID = os.Getenv("IGNITION_CLIENT_ID")
			clientSecret = os.Getenv("IGNITION_CLIENT_SECRET")
			issuerURL = os.Getenv("IGNITION_ISSUER_URL")
			authURL = os.Getenv("IGNITION_AUTH_URL")
			tokenURL = os.Getenv("IGNITION_TOKEN_URL")
			jwksURL = os.Getenv("IGNITION_JWKS_URL")
		})
		it.After(func() {
			os.Setenv("IGNITION_AUTH_VARIANT", authVariant)
			os.Setenv("IGNITION_CLIENT_ID", clientID)
			os.Setenv("IGNITION_CLIENT_SECRET", clientSecret)
			os.Setenv("IGNITION_ISSUER_URL", issuerURL)
			os.Setenv("IGNITION_AUTH_URL", authURL)
			os.Setenv("IGNITION_TOKEN_URL", tokenURL)
			os.Setenv("IGNITION_JWKS_URL", jwksURL)
			os.Unsetenv("VCAP_APPLICATION")
			os.Unsetenv("VCAP_SERVICES")
			os.Unsetenv("PORT")
		})

		when("running outside of cf", func() {
			it.Before(func() {
				root, _ := os.Getwd()
				expectedWebRoot = filepath.Join(root, "web", "dist")
			})

			it("returns config", func() {
				Expect(buildConfig()).NotTo(BeNil())
			})

			it("uses the IGNITION_CLIENT_SECRET environment variable for clientSecret", func() {
				os.Setenv("IGNITION_CLIENT_SECRET", "test-secret")
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.oauth2Config.ClientSecret).To(Equal("test-secret"))
			})

			it("uses the IGNITION_CLIENT_ID environment variable for clientID", func() {
				os.Setenv("IGNITION_CLIENT_ID", "test-ID")
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.oauth2Config.ClientID).To(Equal("test-ID"))
			})

			it("uses the correct webroot", func() {
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.webRoot).To(Equal(expectedWebRoot))
			})

			it("uses the correct port", func() {
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.servePort).To(Equal(3000))
				Expect(c.port).To(Equal(3000))
			})

			it("uses the correct scheme", func() {
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.scheme).To(Equal("http"))
			})

			it("uses the correct domain", func() {
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.domain).To(Equal("localhost"))
			})
		})

		when("running inside of cf with the p-identity variant", func() {
			it.Before(func() {
				os.Setenv("IGNITION_AUTH_VARIANT", "p-identity")
				os.Setenv("VCAP_APPLICATION", `{"cf_api": "https://api.run.pcfbeta.io","limits": {"fds": 16384},"application_name": "ignition","application_uris": ["ignition.pcfbeta.io"],"name": "ignition","space_name": "development","space_id": "test-space-id","uris": ["ignition.pcfbeta.io"],"users": null,"application_id": "test-app-id"}`)
				os.Setenv("VCAP_SERVICES", `{
				  "p-identity": [
				    {
				      "credentials": {
				        "auth_domain": "https://ignition.login.run.pcfbeta.io",
				        "client_secret": "test-cf-client-secret",
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
				os.Setenv("PORT", "12345")
				root, _ := os.Getwd()
				expectedWebRoot = root
			})

			it("returns config", func() {
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c).NotTo(BeNil())
			})

			it("uses the correct port", func() {
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.port).To(Equal(443))
				Expect(c.servePort).To(Equal(12345))
			})

			it("uses the correct webroot", func() {
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.webRoot).To(Equal(expectedWebRoot))
			})

			it("uses the ignition service binding for clientSecret", func() {
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.oauth2Config.ClientSecret).To(Equal("test-cf-client-secret"))
			})

			it("uses the ignition service binding for clientID", func() {
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.oauth2Config.ClientID).To(Equal("test-cf-client-id"))
			})

			it("uses the IGNITION_AUTH_URL env var to get the auth URL", func() {
				os.Setenv("IGNITION_AUTH_URL", "test-auth-uri")
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.oauth2Config.Endpoint.AuthURL).To(Equal("test-auth-uri"))
			})

			it("uses the IGNITION_TOKEN_URL env var to get the token URL", func() {
				os.Setenv("IGNITION_TOKEN_URL", "test-token-uri")
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.oauth2Config.Endpoint.TokenURL).To(Equal("test-token-uri"))
			})

			it("uses the IGNITION_JWKS_URL env var to get the jwks URL", func() {
				os.Setenv("IGNITION_JWKS_URL", "test-jwks-uri")
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.jwksURL).To(Equal("test-jwks-uri"))
			})

			it("uses the IGNITION_ISSUER_URL env var to get the issuer URL", func() {
				os.Setenv("IGNITION_ISSUER_URL", "test-issuer-uri")
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.issuerURL).To(Equal("test-issuer-uri"))
			})

			it("fails if sso instance is not bound with name identity", func() {
				os.Setenv("VCAP_SERVICES", `{
					"p-identity": [
						{
							"credentials": {
								"auth_domain": "https://ignition.login.run.pcfbeta.io",
								"client_secret": "test-cf-client-secret",
								"client_id": "test-cf-client-id"
							},
							"syslog_drain_url": null,
							"volume_mounts": [],
							"label": "p-identity",
							"provider": null,
							"plan": "ignition",
							"name": "a-different-name",
							"tags": []
						}
					]
				}`)

				c, err := buildConfig()
				Expect(c).To(BeNil())
				Expect(err).To(HaveOccurred())
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

				c, err := buildConfig()
				Expect(c).To(BeNil())
				Expect(err).To(HaveOccurred())
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

				c, err := buildConfig()
				Expect(c).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not retrieve the client_secret; make sure you have created and bound a Single Sign On service instance with the name \"identity\""))
			})

			it("uses the correct scheme", func() {
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.scheme).To(Equal("https"))
			})

			it("uses the correct domain", func() {
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.domain).To(Equal("ignition.pcfbeta.io"))
			})
		})

		when("running inside of cf with generic oauth2", func() {
			it.Before(func() {
				os.Setenv("IGNITION_AUTH_VARIANT", "google")
				os.Setenv("VCAP_APPLICATION", `{"cf_api": "https://api.run.pcfbeta.io","limits": {"fds": 16384},"application_name": "ignition","application_uris": ["ignition.pcfbeta.io"],"name": "ignition","space_name": "development","space_id": "test-space-id","uris": ["ignition.pcfbeta.io"],"users": null,"application_id": "test-app-id"}`)
				os.Setenv("VCAP_SERVICES", `{}`)
				os.Setenv("PORT", "12345")
				root, _ := os.Getwd()
				expectedWebRoot = root
			})

			it("returns config", func() {
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c).NotTo(BeNil())
			})

			it("uses the IGNITION_CLIENT_SECRET environment variable for clientSecret", func() {
				os.Setenv("IGNITION_CLIENT_SECRET", "test-secret")
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.oauth2Config.ClientSecret).To(Equal("test-secret"))
			})

			it("uses the IGNITION_CLIENT_ID environment variable for clientID", func() {
				os.Setenv("IGNITION_CLIENT_ID", "test-ID")
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.oauth2Config.ClientID).To(Equal("test-ID"))
			})

			it("uses the IGNITION_AUTH_URL env var to get the auth URL", func() {
				os.Setenv("IGNITION_AUTH_URL", "test-auth-uri")
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.oauth2Config.Endpoint.AuthURL).To(Equal("test-auth-uri"))
			})

			it("uses the IGNITION_TOKEN_URL env var to get the token URL", func() {
				os.Setenv("IGNITION_TOKEN_URL", "test-token-uri")
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.oauth2Config.Endpoint.TokenURL).To(Equal("test-token-uri"))
			})

			it("uses the IGNITION_JWKS_URL env var to get the jwks URL", func() {
				os.Setenv("IGNITION_JWKS_URL", "test-jwks-uri")
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.jwksURL).To(Equal("test-jwks-uri"))
			})

			it("uses the IGNITION_ISSUER_URL env var to get the issuer URL", func() {
				os.Setenv("IGNITION_ISSUER_URL", "test-issuer-uri")
				c, err := buildConfig()
				Expect(err).NotTo(HaveOccurred())
				Expect(c.issuerURL).To(Equal("test-issuer-uri"))
			})
		})
	})
}
