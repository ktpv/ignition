package cloudfoundry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

// Info contains cloud controller API information
type Info struct {
	Name                     string `json:"name"`
	Build                    string `json:"build"`
	Support                  string `json:"support"`
	Version                  int    `json:"version"`
	Description              string `json:"description"`
	AuthorizationEndpoint    string `json:"authorization_endpoint"`
	TokenEndpoint            string `json:"token_endpoint"`
	MinCliVersion            string `json:"min_cli_version"`
	MinRecommendedCliVersion string `json:"min_recommended_cli_version"`
	APIVersion               string `json:"api_version"`
	AppSSHEndpoint           string `json:"app_ssh_endpoint"`
	AppSSHHostKeyFingerprint string `json:"app_ssh_host_key_fingerprint"`
	AppSSHOauthClient        string `json:"app_ssh_oauth_client"`
	DopplerLoggingEndpoint   string `json:"doppler_logging_endpoint"`
	RoutingEndpoint          string `json:"routing_endpoint"`
}

// Info returns the Cloud Controller API info
func (a *API) Info() (*Info, error) {
	if strings.TrimSpace(a.URI) == "" {
		return nil, errors.New("cannot get cloud controller info for empty URI")
	}
	uri := fmt.Sprintf("%s/v2/info", a.URI)
	r, err := http.Get(uri)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("could not get cloud controller info for URI: [%s]", uri))
	}
	var i Info
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&i)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("could not decode response from URI: [%s]", uri))
	}
	return &i, nil
}
