// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package container

import (
	"bufio"
	"bytes"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"io"
	"strconv"
	"strings"
)

type Controller struct {
	cli *client.Client
}

// NewController creates a new Controller struct, which is used to interact
// with a container runtime engine (e.g. Docker or Podman etc.). It initializes
// engine client by calling client.NewClientWithOpts(client.FromEnv) and
// setting the cli field of the Controller struct to the result.
// The Controller struct is then returned.
func NewController() (c *Controller) {
	var err error
	c = new(Controller)
	c.cli, err = client.NewClientWithOpts(client.FromEnv)
	checkErr(err)

	return
}

// EnsureImage creates a new instance of the Controller struct and initializes
// the container engine client by calling client.NewClientWithOpts() with
// the client.FromEnv option. It returns the Controller instance and an error
// if one occurred while creating the client. The Controller struct has a
// method EnsureImage() which pulls an image with the given name from
// a registry and logs the output to the console.
func (c Controller) EnsureImage(image string) (err error) {
	var reader io.ReadCloser

	trace("Running ImagePull for image %s", image)
	reader, err = c.cli.ImagePull(context.Background(), image, types.ImagePullOptions{})
	if reader != nil {
		defer func() {
			checkErr(reader.Close())
		}()

		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			trace(scanner.Text())
		}
	}

	return
}

// ContainerRun creates a new container using the provided image and env values
// and binds it to the specified port number. It then starts the container and returns
// the ID of the container.
func (c Controller) ContainerRun(image string, env []string, port int, command []string, unitTestFailure bool) string {
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port("1433/tcp"): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: strconv.Itoa(port),
				},
			},
		},
	}

	resp, err := c.cli.ContainerCreate(context.Background(), &container.Config{
		Tty:   true,
		Image: image,
		Cmd:   command,
		Env:   env,
	}, hostConfig, nil, nil, "")
	checkErr(err)

	err = c.cli.ContainerStart(
		context.Background(),
		resp.ID,
		types.ContainerStartOptions{},
	)
	if err != nil || unitTestFailure {
		// Remove the container, because we haven't persisted to config yet, so
		// uninstall won't work yet
		if resp.ID != "" {
			err := c.ContainerRemove(resp.ID)
			checkErr(err)
		}
	}
	checkErr(err)

	return resp.ID
}

// ContainerWaitForLogEntry is used to wait for a specific string to be written
// to the logs of a container with the given ID. The function takes in the ID
// of the container and the string to look for in the logs. It creates a reader
// to stream the logs from the container, and scans the logs line by line until
// it finds the specified string. Once the string is found, the function breaks
// out of the loop and returns.
//
// This function is useful for waiting until a specific event has occurred in the
// container (e.g. a server has started up) before continuing with other operations.
func (c Controller) ContainerWaitForLogEntry(id string, text string) {
	options := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: false,
		Since:      "",
		Until:      "",
		Timestamps: false,
		Follow:     true,
		Tail:       "",
		Details:    false,
	}

	if text == "unit-testing" {
		return
	}

	// Wait for server to start up
	reader, err := c.cli.ContainerLogs(context.Background(), id, options)
	checkErr(err)
	defer func() {
		err := reader.Close()
		checkErr(err)
	}()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		trace("ERRORLOG: " + scanner.Text())
		if strings.Contains(scanner.Text(), text) {
			break
		}
	}
}

// ContainerStop stops the container with the given ID. The function returns
// an error if there is an issue stopping the container.
func (c Controller) ContainerStop(id string) (err error) {
	if id == "" {
		panic("Must pass in non-empty id")
	}

	err = c.cli.ContainerStop(context.Background(), id, nil)
	return
}

// ContainerFiles returns a list of files matching a specified pattern within
// a given container. It takes an id argument, which specifies the ID of the
// container to search, and a filespec argument, which is a string pattern used
// to match files within the container. The function returns a []string slice
// containing the names of all files that match the specified pattern.
func (c Controller) ContainerFiles(id string, filespec string) (files []string) {
	if id == "" {
		panic("Must pass in non-empty id")
	}
	if filespec == "" {
		panic("Must pass in non-empty filespec")
	}

	cmd := []string{"find", "/", "-iname", filespec}
	response, err := c.cli.ContainerExecCreate(
		context.Background(),
		id,
		types.ExecConfig{
			AttachStderr: false,
			AttachStdout: true,
			Cmd:          cmd,
		},
	)
	checkErr(err)

	r, err := c.cli.ContainerExecAttach(
		context.Background(),
		response.ID,
		types.ExecStartCheck{},
	)
	checkErr(err)
	defer r.Close()

	// read the output
	var outBuf, errBuf bytes.Buffer
	outputDone := make(chan error)

	go func() {
		// StdCopy de-multiplexes the stream into two buffers
		_, err = stdcopy.StdCopy(&outBuf, &errBuf, r.Reader)
		outputDone <- err
	}()

	err = <-outputDone
	checkErr(err)
	stdout, err := io.ReadAll(&outBuf)
	checkErr(err)

	return strings.Split(string(stdout), "\n")
}

// ContainerExists checks if a container with the given ID exists in the system.
// It does this by using the container runtime API to list all containers and
// filtering by the given ID. If a container with the given ID is found, it
// returns true; otherwise, it returns false.
func (c Controller) ContainerExists(id string) (exists bool) {
	f := filters.NewArgs()
	f.Add(
		"id", id,
	)
	resp, err := c.cli.ContainerList(
		context.Background(),
		types.ContainerListOptions{Filters: f},
	)
	checkErr(err)
	if len(resp) > 0 {
		trace("%v", resp)
		containerStatus := strings.Split(resp[0].Status, " ")
		status := containerStatus[0]
		trace("%v", status)
		exists = true
	}

	return
}

// ContainerRemove removes the container with the specified ID using the
// container runtime API. The function takes the ID of the container to be
// removed as an input argument, and returns an error if one occurs during
// the removal process.
func (c Controller) ContainerRemove(id string) (err error) {
	if id == "" {
		panic("Must pass in non-empty id")
	}

	options := types.ContainerRemoveOptions{
		RemoveVolumes: false,
		RemoveLinks:   false,
		Force:         false,
	}

	err = c.cli.ContainerRemove(context.Background(), id, options)

	return
}
