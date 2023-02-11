// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package http

import (
	"github.com/stretchr/testify/assert"
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
	assert.True(t, UrlExists(ts.URL), "Expected UrlExists to return true for URL %q, but got false", ts.URL)

	// Test case 2: URL exists but returns a non-200 status code
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()
	assert.False(t, UrlExists(ts.URL), "Expected UrlExists to return false for URL %q, but got true", ts.URL)

	assert.False(t, UrlExists("http://invalid.url"), "Expected UrlExists to return false for invalid URL, but got true")
}
