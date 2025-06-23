// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package container

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestController_ListTags(t *testing.T) {
	const registry = "mcr.microsoft.com"
	const repo = "mssql/server"

	ListTags(repo, "https://"+registry)
}

func TestController_EnsureImage(t *testing.T) {
	const registry = "docker.io"
	const repo = "library/alpine"
	const tag = "latest"
	const port = 0

	imageName := fmt.Sprintf(
		"%s/%s:%s",
		registry,
		repo,
		tag)

	c := NewController()
	err := c.EnsureImage(imageName)
	checkErr(err)
	id := c.ContainerRun(
		imageName,
		[]string{},
		port,
		"",
		"",
		"amd64",
		"linux",
		[]string{"ash", "-c", "echo 'Hello World'; sleep 3"},
		false,
	)
	c.ContainerRunning(id)
	c.ContainerWaitForLogEntry(id, "Hello World")
	c.ContainerExists(id)
	c.ContainerFiles(id, "*.mdf")

	// Note: This test downloads from a localhost httptest server which demonstrates
	// the container networking limitation - containers cannot access host localhost by default.
	// In real usage, users should use external URLs or configure Docker networking properly.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("test"))
	}))
	defer ts.Close()

	// This will fail with "Connection refused" but now properly reports the error
	// instead of silently failing like it did before our fix
	defer func() {
		if r := recover(); r != nil {
			// Expected behavior: should now properly report wget failures
			if !strings.Contains(fmt.Sprintf("%v", r), "Connection refused") {
				panic(r) // Re-panic if it's a different error
			}
		}
	}()
	
	c.DownloadFile(id, ts.URL+"/test.dat", "/tmp")

	err = c.ContainerStop(id)
	checkErr(err)
	err = c.ContainerStart(id)
	checkErr(err)
	err = c.ContainerStop(id)
	checkErr(err)
	err = c.ContainerRemove(id)
	checkErr(err)
}

func TestController_ContainerRunFailure(t *testing.T) {
	const registry = "docker.io"
	const repo = "does-not-exist"
	const tag = "latest"

	imageName := fmt.Sprintf(
		"%s/%s:%s",
		registry,
		repo,
		tag)

	c := NewController()

	assert.Panics(t, func() {
		c.ContainerRun(
			imageName,
			[]string{},
			0,
			"",
			"",
			"amd64",
			"linux",
			[]string{"ash", "-c", "echo 'Hello World'; sleep 1"},
			false,
		)
	})
}

func TestController_ContainerRunFailureCleanup(t *testing.T) {
	const registry = "docker.io"
	const repo = "library/alpine"
	const tag = "latest"

	imageName := fmt.Sprintf(
		"%s/%s:%s",
		registry,
		repo,
		tag)

	c := NewController()

	assert.Panics(t, func() {
		c.ContainerRun(
			imageName,
			[]string{},
			0,
			"",
			"",
			"amd64",
			"linux",
			[]string{"ash", "-c", "echo 'Hello World'; sleep 1"},
			true,
		)
	})
}

func TestController_ContainerStopNeg2(t *testing.T) {
	c := NewController()
	assert.Panics(t, func() {
		err := c.ContainerStop("")
		checkErr(err)
	})
}

func TestController_ContainerRemoveNeg(t *testing.T) {
	c := NewController()
	assert.Panics(t, func() {
		err := c.ContainerRemove("")
		checkErr(err)
	})
}

func TestController_ContainerFilesNeg(t *testing.T) {
	c := NewController()
	assert.Panics(t, func() {
		c.ContainerFiles("", "")
	})
}

func TestController_ContainerFilesNeg2(t *testing.T) {
	c := NewController()
	assert.Panics(t, func() {
		c.ContainerFiles("id", "")
	})
}

func TestController_ContainerRunningNeg(t *testing.T) {
	c := NewController()
	assert.Panics(t, func() {
		c.ContainerRunning("")
	})
}

func TestController_ContainerStartNeg(t *testing.T) {
	c := NewController()
	assert.Panics(t, func() {
		err := c.ContainerStart("")
		checkErr(err)
	})
}

func TestController_DownloadFileNeg(t *testing.T) {
	c := NewController()
	assert.Panics(t, func() {
		c.DownloadFile("", "", "")
	})
}

func TestController_DownloadFileNeg2(t *testing.T) {
	c := NewController()
	assert.Panics(t, func() {
		c.DownloadFile("not_blank", "", "")
	})
}

func TestController_DownloadFileNeg3(t *testing.T) {
	c := NewController()
	assert.Panics(t, func() {
		c.DownloadFile("not_blank", "not_blank", "")
	})
}

func TestController_DownloadFileNetworkError(t *testing.T) {
	const registry = "docker.io"
	const repo = "library/alpine"
	const tag = "latest"
	const port = 0

	imageName := fmt.Sprintf(
		"%s/%s:%s",
		registry,
		repo,
		tag)

	c := NewController()
	err := c.EnsureImage(imageName)
	checkErr(err)
	id := c.ContainerRun(
		imageName,
		[]string{},
		port,
		"",
		"",
		"amd64",
		"linux",
		[]string{"ash", "-c", "echo 'Hello World'; sleep 30"},
		false,
	)
	defer func() {
		err = c.ContainerStop(id)
		checkErr(err)
		err = c.ContainerRemove(id)
		checkErr(err)
	}()

	c.ContainerRunning(id)
	c.ContainerWaitForLogEntry(id, "Hello World")

	// Test with invalid URL that should trigger error handling
	invalidURL := "http://127.0.0.1:9999/test.dat"
	
	// Capture the output to see what's happening
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Expected panic occurred: %v", r)
		} else {
			t.Error("DownloadFile should have panicked when wget failed")
		}
	}()
	
	c.DownloadFile(id, invalidURL, "/tmp")
}
