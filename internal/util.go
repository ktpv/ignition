package internal

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

// HelperLoadBytes reads a file from the testdata directory
func HelperLoadBytes(t *testing.T, name string) []byte {
	path := filepath.Join("testdata", name) // relative path
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

// TestDataHandler is a http.Handler that loads the given filename from the
// testdata directory and returns it
func TestDataHandler(t *testing.T, s string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		b := HelperLoadBytes(t, s)
		w.Write(b)
	})
}

// TestDataServer is a test server that uses the TestDataHandler
func TestDataServer(t *testing.T, s string) *httptest.Server {
	return httptest.NewServer(TestDataHandler(t, s))
}
