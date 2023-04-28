package root

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/output"
)

type useOutput struct {
	*output.Output
	output *output.Output
}

func (u *useOutput) FatalNoContainerInCurrentContext() {
	u.output.FatalfWithHintExamples([][]string{
		{"Create a context with a container", "sqlcmd create mssql"},
	}, "Current context does not have a container")
}

func (u *useOutput) FatalContainerNotRunning() {
	u.output.FatalfWithHintExamples([][]string{
		{"Start container for current context", "sqlcmd start"},
	}, "Container for current context is not running")
}

func (u *useOutput) FatalDatabaseSourceFileNotExist(url string) {
	u.output.FatalfWithHints(
		[]string{fmt.Sprintf("File does not exist at URL %q", url)},
		"Unable to download file to container")
}
