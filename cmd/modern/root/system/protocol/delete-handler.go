package protocol

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/cmdparser"
	"github.com/microsoft/go-sqlcmd/internal/pal"
	"runtime"
)

// DeleteHandler implements the `sqlcmd system protocol delete-handler` command
type DeleteHandler struct {
	cmdparser.Cmd
}

func (c *DeleteHandler) DefineCommand(...cmdparser.CommandOptions) {
	examples := []cmdparser.ExampleOptions{
		{
			Description: "Add a user (using the SQLCMD_PASSWORD environment variable)",
			Steps: []string{
				fmt.Sprintf(`%s SQLCMD_PASSWORD=<placeholderpassword>`, pal.CreateEnvVarKeyword()),
				"sqlcmd config add-user --name my-user --username user1 --password-encryption none",
				fmt.Sprintf(`%s SQLCMD_PASSWORD=`, pal.CreateEnvVarKeyword()),
			},
		},
		{
			Description: "Add a user (using the SQLCMDPASSWORD environment variable)",
			Steps: []string{
				fmt.Sprintf(`%s SQLCMDPASSWORD=<placeholderpassword>`, pal.CreateEnvVarKeyword()),
				"sqlcmd config add-user --name my-user --username user1 --password-encryption none",
				fmt.Sprintf(`%s SQLCMDPASSWORD=`, pal.CreateEnvVarKeyword()),
			},
		},
	}

	if runtime.GOOS == "windows" {
		examples = append(examples, cmdparser.ExampleOptions{
			Description: "Add a user using Windows Data Protection API to encrypt password in sqlconfig",
			Steps: []string{
				fmt.Sprintf(`%s SQLCMD_PASSWORD=<placeholderpassword>`, pal.CreateEnvVarKeyword()),
				"sqlcmd config add-user --name my-user --username user1 --password-encryption dpapi",
				fmt.Sprintf(`%s SQLCMD_PASSWORD=`, pal.CreateEnvVarKeyword()),
			},
		})
	}

	options := cmdparser.CommandOptions{
		Use:      "delete-handler",
		Short:    "Delete the sqlcmd:// protocol handler",
		Examples: examples,
		Run:      c.run}

	c.Cmd.DefineCommand(options)

}

func (c *DeleteHandler) run() {
	_ = c.Output()
}
