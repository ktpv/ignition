package http

import (
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Run starts a server listening on the given serveURI
func Run(serveURI string, webroot string) error {
	r := mux.NewRouter()
	r.Handle("/", ensureHTTPS(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, filepath.Join(webroot, "index.html"))
	})))

	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir(path.Join(webroot, "assets")+string(os.PathSeparator)))))
	return http.ListenAndServe(serveURI, handlers.LoggingHandler(os.Stdout, handlers.CORS()(r)))
}
