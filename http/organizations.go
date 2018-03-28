package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/pivotalservices/ignition/user"
)

type Org struct {
	GUID                        string `json:"guid"`
	CreatedAt                   string `json:"created_at"`
	UpdatedAt                   string `json:"updated_at"`
	Name                        string `json:"name"`
	QuotaDefinitionGUID         string `json:"quota_definition_guid"`
	DefaultIsolationSegmentGUID string `json:"default_isolation_segment_guid"`
	URL                         string `json:"url"`
}

type OrgQuerier interface {
	ListOrgsByQuery(query url.Values) ([]cfclient.Org, error)
}

// OrgsForUserID returns the orgs that the user is a member of
func OrgsForUserID(id string, appsURL string, q OrgQuerier) ([]Org, error) {
	query := url.Values{}
	query.Add("q", fmt.Sprintf("user_guid:%s", id))
	o, err := q.ListOrgsByQuery(query)
	if err != nil {
		return nil, err
	}
	result := make([]Org, len(o))
	for i := range o {
		result[i] = Org{
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

func organizationHandler(appsURL string, q OrgQuerier) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		profile, err := user.ProfileFromContext(req.Context())
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		userID, err := UserIDFromContext(req.Context())
		if err != nil || strings.TrimSpace(userID) == "" {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		o, err := OrgsForUserID(userID, appsURL, q)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(o) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if len(o) == 1 {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(o[0])
			return
		}

		expected := orgName(profile.AccountName)
		for i := range o {
			if strings.EqualFold(expected, o[i].Name) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(o[i])
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(o[0])
	}
	return http.HandlerFunc(fn)
}

func orgName(accountName string) string {
	if strings.Contains(accountName, "@") {
		components := strings.Split(accountName, "@")
		return fmt.Sprintf("ignition-%s", components[0])
	}

	if strings.Contains(accountName, "\\") {
		components := strings.Split(accountName, "\\")
		return fmt.Sprintf("ignition-%s", components[1])
	}

	return fmt.Sprintf("ignition-%s", accountName)
}
