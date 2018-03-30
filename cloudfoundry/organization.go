package cloudfoundry

import (
	"fmt"
	"net/url"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
)

// Organization is a Cloud Foundry Organization
type Organization struct {
	GUID                        string `json:"guid"`
	CreatedAt                   string `json:"created_at"`
	UpdatedAt                   string `json:"updated_at"`
	Name                        string `json:"name"`
	QuotaDefinitionGUID         string `json:"quota_definition_guid"`
	DefaultIsolationSegmentGUID string `json:"default_isolation_segment_guid"`
	URL                         string `json:"url"`
}

// API is a Cloud Controller API
type API interface {
	OrganizationQuerier
}

// OrganizationQuerier is used to query a Cloud Controller API or organizations
type OrganizationQuerier interface {
	ListOrgsByQuery(query url.Values) ([]cfclient.Org, error)
}

// OrgsForUserID returns the orgs that the user is a member of
func OrgsForUserID(id string, appsURL string, q OrganizationQuerier) ([]Organization, error) {
	query := url.Values{}
	query.Add("q", fmt.Sprintf("user_guid:%s", id))
	o, err := q.ListOrgsByQuery(query)
	if err != nil {
		return nil, err
	}
	result := make([]Organization, len(o))
	for i := range o {
		result[i] = Organization{
			GUID:                        o[i].Guid,
			CreatedAt:                   o[i].CreatedAt,
			UpdatedAt:                   o[i].UpdatedAt,
			Name:                        o[i].Name,
			QuotaDefinitionGUID:         o[i].QuotaDefinitionGuid,
			DefaultIsolationSegmentGUID: o[i].DefaultIsolationSegmentGuid,
			URL: fmt.Sprintf("%s/organizations/%s", appsURL, o[i].Guid),
		}
	}
	return result, nil
}
