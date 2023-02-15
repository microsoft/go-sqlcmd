// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package http

import "net/http"

func UrlExists(url string) (exists bool) {
	trace("http.Head to %q", url)
	resp, err := http.Head(url)
	if err != nil {
		trace("http.Head to %q failed with %v", url, err)
		return false
	}
	if resp.StatusCode != 200 {
		trace("http.Head to %q returned status code %d", url, resp.StatusCode)
		return false
	}

	trace("http.Head to %q succeeded", url)

	return true
}
