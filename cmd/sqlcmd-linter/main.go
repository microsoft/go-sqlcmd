package main

import (
	sqlcmdlinter "github.com/microsoft/go-sqlcmd/pkg/sqlcmd-linter"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(sqlcmdlinter.AssertAnalyzer, sqlcmdlinter.ImportsAnalyzer)
}
