// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package container

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
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
<<<<<<< HEAD

	runOptions := RunOptions{
		Env:          []string{},
		Port:         port,
		Architecture: "amd64",
		Os:           "linux",
		Command:      []string{"ash", "-c", "echo 'Hello World'; sleep 3"},
	}

	id := c.ContainerRun(imageName, runOptions)
=======
	id := c.ContainerRun(imageName, []string{}, nil, 1433, port, "", "", "amd64", "linux", "", []string{"ash", "-c", "echo 'Hello World'; sleep 3"}, false)
>>>>>>> stuartpa/add-ons
	c.ContainerRunning(id)
	c.ContainerWaitForLogEntry(id, "Hello World")
	c.ContainerExists(id)
	c.ContainerFiles(id, "*.mdf")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("test"))
	}))
	defer ts.Close()

	c.DownloadFile(id, ts.URL, "test.txt")

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

<<<<<<< HEAD
	runOptions := RunOptions{
		Architecture: "amd64",
		Os:           "linux",
		Command:      []string{"ash", "-c", "echo 'Hello World'; sleep 1"},
	}

	assert.Panics(t, func() { c.ContainerRun(imageName, runOptions) })
=======
	assert.Panics(t, func() {
		c.ContainerRun(imageName, []string{}, nil, 0, 0, "", "", "amd64", "linux", "", []string{"ash", "-c", "echo 'Hello World'; sleep 1"}, false)
	})
>>>>>>> stuartpa/add-ons
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

<<<<<<< HEAD
	runOptions := RunOptions{
		Architecture:    "amd64",
		Os:              "linux",
		Command:         []string{"ash", "-c", "echo 'Hello World'; sleep 1"},
		UnitTestFailure: true,
	}
	assert.Panics(t, func() { c.ContainerRun(imageName, runOptions) })
=======
	assert.Panics(t, func() {
		c.ContainerRun(imageName, []string{}, nil, 0, 0, "", "", "amd64", "linux", "", []string{"ash", "-c", "echo 'Hello World'; sleep 1"}, true)
	})
>>>>>>> stuartpa/add-ons
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
