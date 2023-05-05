package root

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/output"
	"runtime"
)

type useOutput struct {
	output.Output
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

func (u *useOutput) InfoDatabaseOnline(databaseName string) {
	hints := [][]string{}

	// TODO: sqlcmd open ads only support on Windows/Mac right now, add Linux support
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		hints = append(hints, []string{"Open in Azure Data Studio", "sqlcmd open ads"})
	}

	hints = append(hints, []string{"Run a query", "sqlcmd query \"SELECT DB_NAME()\""})
	hints = append(hints, []string{"See connection strings", "sqlcmd config connection-strings"})

	u.output.InfofWithHintExamples(hints, "Database %q is now online", databaseName)
}
