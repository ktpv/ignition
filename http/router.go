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
func Run(serveURI string) error {
	r := mux.NewRouter()
	root, _ := os.Getwd()
	webRoot := filepath.Join(root, "web")
	r.Handle("/", ensureHTTPS(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, filepath.Join(webRoot, "dist", "index.html"))
	})))

	r.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir(path.Join(webRoot, "dist", "assets")+string(os.PathSeparator)))))
	return http.ListenAndServe(serveURI, handlers.LoggingHandler(os.Stdout, handlers.CORS()(r)))
}
