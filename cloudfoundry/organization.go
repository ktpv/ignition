package cloudfoundry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type ApiOrganization struct {
	NumberOfOrgs int            `json:"total_results"`
	Res          []ApiResources `json:"resources"`
}

type ApiResources struct {
	Metadata ApiMetadata `json:"metadata"`
	Entity   ApiEntity   `json:"entity"`
}

type ApiMetadata struct {
	Guid string `json:"guid"`
}

type ApiEntity struct {
	Name string `json:"name"`
}

type Org struct {
	Guid string
	Name string
}

// Organization returns the Cloud Controller API organizations
func (a *API) Orgs() (*[]Org, error) {
	if strings.TrimSpace(a.URI) == "" {
		return nil, errors.New("cannot get cloud controller orgs for empty URI")
	}
	uri := fmt.Sprintf("%s/v2/organizations", a.URI)
	r, err := http.Get(uri)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("could not get cloud controller orgs for URI: [%s]", uri))
	}
	var o ApiOrganization
	defer r.Body.Close()
	err = json.NewDecoder(r.Body).Decode(&o)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("could not decode response from URI: [%s]", uri))
	}

	var orgs = []Org{
		Org{
			Guid: o.Res[0].Metadata.Guid,
			Name: o.Res[0].Entity.Name,
		},
	}

	return &orgs, nil
}
