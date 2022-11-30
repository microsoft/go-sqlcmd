// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package root

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/config"
	"github.com/microsoft/go-sqlcmd/internal/container"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"path/filepath"
	"strings"
)

type Uninstall struct {
	cmdparser.Cmd

	force bool
	yes   bool
}

// systemDatabases are the list of non-user databases, used to do a safety check
// when doing a delete/drop/uninstall
var systemDatabases = [...]string{
	"/var/opt/mssql/data/msdbdata.mdf",
	"/var/opt/mssql/data/tempdb.mdf",
	"/var/opt/mssql/data/model.mdf",
	"/var/opt/mssql/data/model_msdbdata.mdf",
	"/var/opt/mssql/data/model_replicatedmaster.mdf",
	"/var/opt/mssql/data/master.mdf",
}

func (c *Uninstall) DefineCommand(output.Output, ...cmdparser.Command) {
	c.Cmd.SetOptions(cmdparser.Options{
		Use:   "uninstall",
		Short: "Uninstall/Delete the current context",
		Examples: []cmdparser.ExampleInfo{
			{
				Description: "Uninstall/Delete the current context (includes the endpoint and user)",
				Steps:       []string{`sqlcmd uninstall`}},
			{
				Description: "Uninstall/Delete the current context, no user prompt",
				Steps:       []string{`sqlcmd uninstall --yes`}},
			{
				Description: "Uninstall/Delete the current context, no user prompt and override safety check for user databases",
				Steps:       []string{`sqlcmd uninstall --yes --force`}},
		},
		Aliases: []string{"delete", "drop"},
		Run:     c.run,
	})

	c.Cmd.DefineCommand()

	c.AddFlag(cmdparser.FlagOptions{
		Bool:  &c.yes,
		Name:  "yes",
		Usage: "Quiet mode (do not stop for user input to confirm the operation)",
	})

	c.AddFlag(cmdparser.FlagOptions{
		Bool:  &c.force,
		Name:  "force",
		Usage: "Complete the operation even if non-system (user) database files are present",
	})
}

func (c *Uninstall) run() {
	output := c.Output()

	if config.GetCurrentContextName() == "" {
		output.FatalfWithHintExamples([][]string{
			{"To view available contexts", "sqlcmd config get-contexts"},
		}, "No current context")
	}
	if c.currentContextEndPointExists() {
		if config.CurrentContextEndpointHasContainer() {
			controller := container.NewController()
			id := config.GetContainerId()
			endpoint, _ := config.GetCurrentContext()

			var input string
			if !c.yes {
				output.Infof(
					"Current context is %q. Do you want to continue? (Y/N)",
					config.GetCurrentContextName(),
				)
				_, err := fmt.Scanln(&input)
				c.CheckErr(err)

				if strings.ToLower(input) != "yes" && strings.ToLower(input) != "y" {
					output.Fatal("Operation cancelled.")
				}
			}
			if !c.force {
				output.Infof("Verifying no user (non-system) database (.mdf) files")
				c.userDatabaseSafetyCheck(controller, id)
			}

			output.Infof(
				"Stopping %s",
				endpoint.ContainerDetails.Image,
			)
			err := controller.ContainerStop(id)
			c.CheckErr(err)

			output.Infof("Removing context %s", config.GetCurrentContextName())
			err = controller.ContainerRemove(id)
			c.CheckErr(err)
		}

		config.RemoveCurrentContext()
		config.Save()

		newContextName := config.GetCurrentContextName()
		if newContextName != "" {
			output.Infof("Current context is now %s", newContextName)
		} else {
			output.Infof("%v", "Operation completed successfully")
		}
	}
}

func (c *Uninstall) userDatabaseSafetyCheck(controller *container.Controller, id string) {
	output := c.Output()
	files := controller.ContainerFiles(id, "*.mdf")
	for _, databaseFile := range files {
		if strings.HasSuffix(databaseFile, ".mdf") {
			isSystemDatabase := false
			for _, systemDatabase := range systemDatabases {
				if databaseFile == systemDatabase {
					isSystemDatabase = true
					break
				}
			}

			if !isSystemDatabase {
				output.FatalfWithHints([]string{
					fmt.Sprintf(
						"If the database is mounted, run `sqlcmd query \"use master; DROP DATABASE [%s]\"`",
						strings.TrimSuffix(filepath.Base(databaseFile), ".mdf")),
					"Pass in the flag --force to override this safety check for user (non-system) databases"},
					"Unable to continue, a user (non-system) database (%s) is present", databaseFile)
			}
		}
	}
}

func (c *Uninstall) currentContextEndPointExists() (exists bool) {
	output := c.Output()
	exists = true

	if !config.EndpointsExists() {
		output.Fatal("No endpoints to uninstall")
		exists = false
	}

	return
}
