// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package container

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"os"
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

func (c Controller) NetworkCreate(name string) string {
	resp, err := c.cli.NetworkCreate(context.Background(), name, types.NetworkCreate{})
	checkErr(err)

	return resp.ID
}

func (c Controller) NetworkDelete(name string) {
	err := c.cli.NetworkRemove(context.Background(), name)
	checkErr(err)
}

func (c Controller) NetworkExists(name string) bool {
	networks, err := c.cli.NetworkList(context.Background(), types.NetworkListOptions{})
	checkErr(err)

	for _, network := range networks {
		if network.Name == name {
			return true
		}
	}

	return false
}

// ContainerRun creates a new container using the provided image and env values
// and binds the internal port to the specified external port number. It then starts
// the container and returns the ID of the container.
func (c Controller) ContainerRun(
	image string,
	options RunOptions,
) string {
	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(strconv.Itoa(options.PortInternal) + "/tcp"): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: strconv.Itoa(options.Port),
				},
			},
		},
		NetworkMode: container.NetworkMode(options.Network),
	}

	platform := specs.Platform{
		Architecture: options.Architecture,
		OS:           options.Os,
	}

	resp, err := c.cli.ContainerCreate(context.Background(), &container.Config{
		Tty:      true,
		Image:    image,
		Cmd:      options.Command,
		Env:      options.Env,
		Hostname: options.Hostname,
		ExposedPorts: nat.PortSet{
			nat.Port(strconv.Itoa(options.PortInternal) + "/tcp"): {},
		},
	}, hostConfig, nil, &platform, options.Name)
	checkErr(err)

	err = c.cli.ContainerStart(
		context.Background(),
		resp.ID,
		types.ContainerStartOptions{},
	)
	if err != nil || options.UnitTestFailure {
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

func (c Controller) ContainerName(containerID string) string {
	// Inspect the container to get details
	containerInfo, err := c.cli.ContainerInspect(context.Background(), containerID)
	checkErr(err)

	// Access the container name from the inspect result
	containerName := containerInfo.Name[1:] // Removing the leading '/'
	return containerName
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

	err = c.cli.ContainerStop(context.Background(), id, container.StopOptions{})
	return
}

// ContainerStart starts the container with the given ID. The function returns
// an error if there is an issue starting the container.
func (c Controller) ContainerStart(id string) (err error) {
	if id == "" {
		panic("Must pass in non-empty id")
	}

	err = c.cli.ContainerStart(context.Background(), id, types.ContainerStartOptions{})
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

func (c Controller) CopyFile(id string, src string, destFolder string) {
	if id == "" {
		panic("Must pass in non-empty id")
	}
	if src == "" {
		panic("Must pass in non-empty src")
	}
	if destFolder == "" {
		panic("Must pass in non-empty destFolder")
	}

	trace("Copying file %s to %s", src, destFolder)

	_, f := filepath.Split(src)
	h, err := os.ReadFile(src)
	checkErr(err)

	// Create and add some files to the archive.
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	defer func() {
		checkErr(tw.Close())
	}()
	hdr := &tar.Header{
		Name: f,
		Mode: 0600,
		Size: int64(len(h)),
	}
	err = tw.WriteHeader(hdr)
	checkErr(err)
	_, err = tw.Write([]byte(h))
	checkErr(err)

	err = c.cli.CopyToContainer(context.Background(), id, destFolder, &buf, types.CopyToContainerOptions{})
	checkErr(err)
}

func (c Controller) DownloadFile(id string, src string, destFolder string) {
	if id == "" {
		panic("Must pass in non-empty id")
	}
	if src == "" {
		panic("Must pass in non-empty src")
	}
	if destFolder == "" {
		panic("Must pass in non-empty destFolder")
	}

	trace("Downloading file %s to %s (will try wget first, and curl if wget fails", src, destFolder)

	cmd := []string{"mkdir", "-p", destFolder}
	c.RunCmdInContainer(id, cmd, ExecOptions{})

	_, file := filepath.Split(src)

	// Wget the .bak file from the http src, and place it in /var/opt/sql/backup
	cmd = []string{
		"wget",
		"-O",
		destFolder + "/" + file, // not using filepath.Join here, this is in the *nix container. always /
		src,
	}

	_, _, exitCode := c.RunCmdInContainer(id, cmd, ExecOptions{})
	trace("wget exit code: %d", exitCode)

	if exitCode == 126 {
		trace("wget was not found in container, trying curl")
		cmd = []string{
			"curl",
			"-o",
			destFolder + "/" + file, // not using filepath.Join here, this is in the *nix container. always /
			"-L",
			src,
		}

		_, _, exitCode = c.RunCmdInContainer(id, cmd, ExecOptions{})
		trace("curl exit code: %d", exitCode)
	}
}

type ExecOptions struct {
	User string
	Env  []string
}

func (c Controller) RunCmdInContainer(id string, cmd []string, options ExecOptions) ([]byte, []byte, int) {
	trace("Running command in container: " + strings.Replace(strings.Join(cmd, " "), "%", "%%", -1))

	response, err := c.cli.ContainerExecCreate(
		context.Background(),
		id,
		types.ExecConfig{
			User:         options.User,
			AttachStderr: true,
			AttachStdout: true,
			Cmd:          cmd,
			Env:          options.Env,
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
	stderr, err := io.ReadAll(&errBuf)
	checkErr(err)

	trace("Stdout: " + strings.Replace(string(stdout), "%", "%%%%", -1))
	trace("Stderr: " + strings.Replace(string(stderr), "%", "%%%%", -1))

	// Get the exit code
	execInspect, err := c.cli.ContainerExecInspect(context.Background(), response.ID)
	checkErr(err)

	trace("ExitCode: %d", execInspect.ExitCode)

	return stdout, stderr, execInspect.ExitCode
}

// ContainerRunning returns true if the container with the given ID is running.
// It returns false if the container is not running or if there is an issue
// getting the container's status.
func (c Controller) ContainerRunning(id string) (running bool) {
	if id == "" {
		panic("Must pass in non-empty id")
	}

	resp, err := c.cli.ContainerInspect(context.Background(), id)
	checkErr(err)
	running = resp.State.Running
	return
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
		types.ContainerListOptions{Filters: f, All: true},
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
