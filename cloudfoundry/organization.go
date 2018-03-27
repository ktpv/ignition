package cloudfoundry

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type orgResult struct {
	TotalResults int         `json:"total_results"`
	TotalPages   int         `json:"total_pages"`
	PrevURL      interface{} `json:"prev_url"`
	NextURL      string      `json:"next_url"`
	Resources    []struct {
		Metadata struct {
			GUID      string    `json:"guid"`
			URL       string    `json:"url"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"metadata"`
		Entity struct {
			Name                        string `json:"name"`
			BillingEnabled              bool   `json:"billing_enabled"`
			QuotaDefinitionGUID         string `json:"quota_definition_guid"`
			Status                      string `json:"status"`
			DefaultIsolationSegmentGUID string `json:"default_isolation_segment_guid"`
			IsolationSegmentURL         string `json:"isolation_segment_url"`
			QuotaDefinitionURL          string `json:"quota_definition_url"`
			SpacesURL                   string `json:"spaces_url"`
			DomainsURL                  string `json:"domains_url"`
			PrivateDomainsURL           string `json:"private_domains_url"`
			UsersURL                    string `json:"users_url"`
			ManagersURL                 string `json:"managers_url"`
			BillingManagersURL          string `json:"billing_managers_url"`
			AuditorsURL                 string `json:"auditors_url"`
			AppEventsURL                string `json:"app_events_url"`
			SpaceQuotaDefinitionsURL    string `json:"space_quota_definitions_url"`
		} `json:"entity"`
	} `json:"resources"`
}

// Organization is a Cloud Foundry Organization
type Organization struct {
	GUID                string `json:"guid"`
	URL                 string `json:"url"`
	Name                string `json:"name"`
	QuotaDefinitionGUID string `json:"quota_definition_guid"`
	Status              string `json:"status"`
	QuotaDefinitionURL  string `json:"quota_definition_url"`
}

// Organizations returns the Cloud Controller API organizations
func (a *API) Organizations(ctx context.Context) ([]Organization, error) {
	err := a.ensureValidAuthentication()
	if err != nil {
		return nil, errors.Wrap(err, "could not authenticate")
	}
	uri := fmt.Sprintf("%s/v2/organizations", a.URI)
	r, err := a.Config.Client(ctx, a.Token).Get(uri)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("could not get cloud controller orgs for URI: [%s]", uri))
	}
	var o orgResult
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&o)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("could not decode response from URI: [%s]", uri))
	}

	result := make([]Organization, len(o.Resources))

	for i := range o.Resources {
		result[i] = Organization{
			GUID:                o.Resources[i].Metadata.GUID,
			URL:                 o.Resources[i].Metadata.URL,
			Name:                o.Resources[i].Entity.Name,
			QuotaDefinitionGUID: o.Resources[i].Entity.QuotaDefinitionGUID,
			QuotaDefinitionURL:  o.Resources[i].Entity.QuotaDefinitionURL,
			Status:              o.Resources[i].Entity.Status,
		}
	}

	return result, nil
}
