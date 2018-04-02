package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/pivotalservices/ignition/cloudfoundry"
	"github.com/pivotalservices/ignition/http/session"
	"github.com/pivotalservices/ignition/user"
)

func organizationHandler(appsURL string, orgPrefix string, quotaID string, q cloudfoundry.OrganizationQuerier) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		profile, err := user.ProfileFromContext(req.Context())
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		userID, err := session.UserIDFromContext(req.Context())
		if err != nil || strings.TrimSpace(userID) == "" {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		o, err := cloudfoundry.OrgsForUserID(userID, appsURL, q)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(o) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		expected := orgName(orgPrefix, profile.AccountName)
		var quotaMatches []cloudfoundry.Organization
		for i := range o {
			if strings.EqualFold(quotaID, o[i].QuotaDefinitionGUID) {
				quotaMatches = append(quotaMatches, o[i])
			}
			if strings.EqualFold(expected, o[i].Name) {
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(o[i])
				return
			}
		}

		if len(quotaMatches) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(quotaMatches[0])
	}
	return http.HandlerFunc(fn)
}

func orgName(orgPrefix string, accountName string) string {
	orgPrefix = strings.ToLower(orgPrefix)
	accountName = strings.ToLower(accountName)
	if strings.Contains(accountName, "@") {
		components := strings.Split(accountName, "@")
		return fmt.Sprintf("%s-%s", orgPrefix, components[0])
	}

	if strings.Contains(accountName, "\\") {
		components := strings.Split(accountName, "\\")
		return fmt.Sprintf("%s-%s", orgPrefix, components[1])
	}

	return fmt.Sprintf("%s-%s", orgPrefix, accountName)
}
