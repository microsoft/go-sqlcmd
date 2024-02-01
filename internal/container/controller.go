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

	_, file := filepath.Split(strings.Split(src, "?")[0])

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
	trace("ContainerExists: " + id)

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

	trace("ContainerExists: %v", exists)

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

// DAB project file
/*
dabCsproj := `<Project Sdk="Microsoft.NET.Sdk.Web">

  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
  </PropertyGroup>`

dabDockerfile := `FROM mcr.microsoft.com/azure-databases/data-api-builder:latest

COPY dab-config.json /App
WORKDIR /App
ENV ASPNETCORE_URLS=http://+:5000
EXPOSE 5000
ENTRYPOINT ["dotnet", "Azure.DataApiBuilder.Service.dll"]
`

// C:\Users\stuartpa\classroom-assignment\infra\core\database\sqlserver\sqlserver.bicep
sqlServerBicep := `metadata description = 'Creates an Azure SQL Server instance.'
param name string
param location string = resourceGroup().location
param tags object = {}

param appUser string = 'appUser'
param databaseName string
param keyVaultName string
param sqlAdmin string = 'sqlAdmin'
param connectionStringKey string = 'AZURE-SQL-CONNECTION-STRING'

@secure()
param sqlAdminPassword string
@secure()
param appUserPassword string

resource sqlServer 'Microsoft.Sql/servers@2022-05-01-preview' = {
  name: name
  location: location
  tags: tags
  properties: {
    version: '12.0'
    minimalTlsVersion: '1.2'
    publicNetworkAccess: 'Enabled'
    administratorLogin: sqlAdmin
    administratorLoginPassword: sqlAdminPassword
  }

  resource database 'databases' = {
    name: databaseName
    location: location
  }

  resource firewall 'firewallRules' = {
    name: 'Azure Services'
    properties: {
      // Allow all clients
      // Note: range [0.0.0.0-0.0.0.0] means "allow all Azure-hosted clients only".
      // This is not sufficient, because we also want to allow direct access from developer machine, for debugging purposes.
      startIpAddress: '0.0.0.1'
      endIpAddress: '255.255.255.254'
    }
  }
}

resource sqlDeploymentScript 'Microsoft.Resources/deploymentScripts@2020-10-01' = {
  name: '${name}-deployment-script'
  location: location
  kind: 'AzureCLI'
  properties: {
    azCliVersion: '2.37.0'
    retentionInterval: 'PT1H' // Retain the script resource for 1 hour after it ends running
    timeout: 'PT5M' // Five minutes
    cleanupPreference: 'OnSuccess'
    environmentVariables: [
      {
        name: 'APPUSERNAME'
        value: appUser
      }
      {
        name: 'APPUSERPASSWORD'
        secureValue: appUserPassword
      }
      {
        name: 'DBNAME'
        value: databaseName
      }
      {
        name: 'DBSERVER'
        value: sqlServer.properties.fullyQualifiedDomainName
      }
      {
        name: 'SQLCMDPASSWORD'
        secureValue: sqlAdminPassword
      }
      {
        name: 'SQLADMIN'
        value: sqlAdmin
      }
    ]

    scriptContent: '''
wget https://github.com/microsoft/go-sqlcmd/releases/download/v0.8.1/sqlcmd-v0.8.1-linux-x64.tar.bz2
tar x -f sqlcmd-v0.8.1-linux-x64.tar.bz2 -C .

cat <<SCRIPT_END > ./initDb.sql
drop user if exists ${APPUSERNAME}
go
create user ${APPUSERNAME} with password = '${APPUSERPASSWORD}'
go
alter role db_owner add member ${APPUSERNAME}
go
SCRIPT_END

./sqlcmd -S ${DBSERVER} -d ${DBNAME} -U ${SQLADMIN} -i ./initDb.sql
    '''
  }
}

resource sqlAdminPasswordSecret 'Microsoft.KeyVault/vaults/secrets@2022-07-01' = {
  parent: keyVault
  name: 'sqlAdminPassword'
  properties: {
    value: sqlAdminPassword
  }
}

resource appUserPasswordSecret 'Microsoft.KeyVault/vaults/secrets@2022-07-01' = {
  parent: keyVault
  name: 'appUserPassword'
  properties: {
    value: appUserPassword
  }
}

resource sqlAzureConnectionStringSercret 'Microsoft.KeyVault/vaults/secrets@2022-07-01' = {
  parent: keyVault
  name: connectionStringKey
  properties: {
    value: '${connectionString}; Password=${appUserPassword}'
  }
}

resource keyVault 'Microsoft.KeyVault/vaults@2022-07-01' existing = {
  name: keyVaultName
}

var connectionString = 'Server=${sqlServer.properties.fullyQualifiedDomainName}; Database=${sqlServer::database.name}; User=${appUser}'
output connectionStringKey string = connectionStringKey
output databaseName string = sqlServer::database.name
`

mainBicepDatabase := `
param sqlDatabaseName string = ''
param sqlServerName string = ''

@secure()
@description('SQL Server administrator password')
param sqlAdminPassword string

@secure()
@description('Application user password')
param appUserPassword string

// The application database
module sqlServer './app/db.bicep' = {
  name: 'sql'
  scope: rg
  params: {
    name: !empty(sqlServerName) ? sqlServerName : '${abbrs.sqlServers}${resourceToken}'
    databaseName: sqlDatabaseName
    location: location
    tags: tags
    sqlAdminPassword: sqlAdminPassword
    appUserPassword: appUserPassword
    keyVaultName: keyVault.outputs.name
  }
}

output AZURE_SQL_CONNECTION_STRING_KEY string = sqlServer.outputs.connectionStringKey
`

// C:\Users\stuartpa\classroom-assignment\infra\app\db.bicep
dbBicep := `param name string
param location string = resourceGroup().location
param tags object = {}

param databaseName string = ''
param keyVaultName string

@secure()
param sqlAdminPassword string
@secure()
param appUserPassword string

// Because databaseName is optional in main.bicep, we make sure the database name is set here.
var defaultDatabaseName = 'sample'
var actualDatabaseName = !empty(databaseName) ? databaseName : defaultDatabaseName

module sqlServer '../core/database/sqlserver/sqlserver.bicep' = {
  name: 'sqlserver'
  params: {
    name: name
    location: location
    tags: tags
    databaseName: actualDatabaseName
    keyVaultName: keyVaultName
    sqlAdminPassword: sqlAdminPassword
    appUserPassword: appUserPassword
  }
}

output connectionStringKey string = sqlServer.outputs.connectionStringKey
output databaseName string = sqlServer.outputs.databaseName
`

*/
