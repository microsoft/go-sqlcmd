// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package container

import (
	"fmt"
	"github.com/docker/docker/client"
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

	type fields struct {
		cli *client.Client
	}
	type args struct {
		image string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"default", fields{nil}, args{imageName}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// If test name ends in 'Panic' expect a Panic
			if strings.HasSuffix(tt.name, "Panic") {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("The code did not panic")
					}
				}()
			}

			c := NewController()
			err := c.EnsureImage(tt.args.image)
			checkErr(err)
			id := c.ContainerRun(tt.args.image, []string{}, port, []string{"ash", "-c", "echo 'Hello World'; sleep 1"}, false)
			c.ContainerWaitForLogEntry(id, "Hello World")
			c.ContainerExists(id)
			c.ContainerFiles(id, "*.mdf")
			err = c.ContainerStop(id)
			checkErr(err)
			err = c.ContainerRemove(id)
			checkErr(err)
		})
	}
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

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	c := NewController()
	c.ContainerRun(
		imageName,
		[]string{},
		0,
		[]string{"ash", "-c", "echo 'Hello World'; sleep 1"},
		false,
	)
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

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	c := NewController()
	c.ContainerRun(
		imageName,
		[]string{},
		0,
		[]string{"ash", "-c", "echo 'Hello World'; sleep 1"},
		true,
	)
}

func TestController_ContainerStopNeg(t *testing.T) {
	const registry = "docker.io"
	const repo = "does-not-exist"
	const tag = "latest"

	imageName := fmt.Sprintf(
		"%s/%s:%s",
		registry,
		repo,
		tag)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	c := NewController()
	c.ContainerRun(imageName, []string{}, 0, []string{"ash", "-c", "echo 'Hello World'; sleep 1"}, false)
}

func TestController_ContainerStopNeg2(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	c := NewController()
	err := c.ContainerStop("")
	checkErr(err)
}

func TestController_ContainerRemoveNeg(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	c := NewController()
	err := c.ContainerRemove("")
	checkErr(err)
}

func TestController_ContainerFilesNeg(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	c := NewController()
	c.ContainerFiles("", "")
}

func TestController_ContainerFilesNeg2(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	c := NewController()
	c.ContainerFiles("id", "")
}
