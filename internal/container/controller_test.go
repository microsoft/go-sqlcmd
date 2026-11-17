// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package container

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
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

	// Bind to 0.0.0.0 so the container can reach the server via the
	// Docker bridge network (host.docker.internal resolves to 172.17.0.1).
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("test"))
	}))
	l, err := net.Listen("tcp4", "0.0.0.0:0")
	checkErr(err)
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	// Build URL from listener port so it works regardless of whether
	// the OS returns 127.0.0.1, localhost, or [::] in ts.URL.
	_, tsPort, _ := net.SplitHostPort(ts.Listener.Addr().String())
	tsURL := fmt.Sprintf("http://host.docker.internal:%s", tsPort)

	c.DownloadFile(id, tsURL+"/test.bak", "/tmp")

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

func TestController_DownloadFileNoFilename(t *testing.T) {
	c := NewController()
	assert.Panics(t, func() {
		c.DownloadFile("not_blank", "http://host:9999/", "/tmp")
	})
}
