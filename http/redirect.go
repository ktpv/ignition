package http

import (
	"fmt"
	"net/http"
)

func ensureHTTPS(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("X-Forwarded-Proto") != "https" && req.Host != "localhost:3000" && req.TLS == nil {
			uri := fmt.Sprintf("https://%v%v", req.Host, req.RequestURI)
			http.Redirect(w, req, uri, http.StatusMovedPermanently)
			return
		}
		next.ServeHTTP(w, req)
	}
	return http.HandlerFunc(fn)
}
