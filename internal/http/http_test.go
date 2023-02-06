package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUrlExists(t *testing.T) {
	// Test case 1: URL exists and returns a 200 status code
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	if !UrlExists(ts.URL) {
		t.Errorf("Expected UrlExists to return true for URL %q, but got false", ts.URL)
	}

	// Test case 2: URL exists but returns a non-200 status code
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	if UrlExists(ts.URL) {
		t.Errorf("Expected UrlExists to return false for URL %q, but got true", ts.URL)
	}

	// Test case 3: URL does not exist
	if UrlExists("http://invalid.url") {
		t.Error("Expected UrlExists to return false for invalid URL, but got true")
	}
}
