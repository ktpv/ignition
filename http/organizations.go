package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func (a *API) organizationHandler() http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		t, err := TokenFromContext(req.Context())
		if err != nil || !t.Valid() {
			log.Println(err)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		o, err := a.CCAPI.Organizations(req.Context())
		fmt.Println(o)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(o)
	}
	return http.HandlerFunc(fn)
}
