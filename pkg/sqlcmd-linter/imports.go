// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmdlinter

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var ImportsAnalyzer = &analysis.Analyzer{
	Name:     "importslint",
	Doc:      "Require most external packages be imported only by internal packages",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runImports,
}

var AllowedImports = map[string][]string{
	`"github.com/alecthomas/kong`:      {`cmd/sqlcmd`, `pkg/sqlcmd`},
	`"github.com/golang-sql/sqlexp`:    {`pkg/sqlcmd`},
	`"github.com/google/uuid`:          {},
	`"github.com/peterh/liner`:         {`pkg/console`},
	`"github.com/microsoft/go-mssqldb`: {},
	`"github.com/microsoft/go-sqlcmd`:  {},
	`"github.com/spf13/cobra`:          {`cmd/sqlcmd`, `cmd/modern`},
	`"github.com/spf13/viper`:          {`cmd/sqlcmd`, `cmd/modern`},
	`"github.com/stretchr/testify`:     {},
}

func runImports(pass *analysis.Pass) (interface{}, error) {
	inspectorInstance := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{(*ast.File)(nil)}
	inspectorInstance.Preorder(nodeFilter, func(n ast.Node) {

		f := n.(*ast.File)
		fileName := pass.Fset.Position(f.Package).Filename
		isInternal := strings.Contains(fileName, `internal\`)
		for _, s := range f.Imports {
			if s.Path.Kind == token.STRING {
				pkg := s.Path.Value
				if isInternal {
					if !isValidInternalImport(pkg) {
						pass.Reportf(s.Pos(), "Internal packages should not import %s", pkg)
					}
				} else if !isValidExternalImport(pkg, fileName) {
					pass.Reportf(s.Pos(), "Non-internal packages should not import %s", pkg)
				}
			}
		}
	})

	return nil, nil
}

func isValidInternalImport(pkg string) bool {
	return !strings.HasPrefix(pkg, `"github.com/microsoft/go-sqlcmd/pkg`) && !strings.HasPrefix(pkg, `"github.com/microsoft/go-sqlcmd/cmd`)
}

func isValidExternalImport(pkg string, filename string) bool {
	if strings.HasPrefix(pkg, `"github.com`) {
		for key, paths := range AllowedImports {
			if strings.HasPrefix(pkg, key) {
				if len(paths) == 0 {
					// any package can import it
					return true
				}
				for _, p := range paths {
					// canonicalize path to Linux separator for comparison
					path := strings.ReplaceAll(filename, `\`, `/`)
					if strings.Contains(path, p) {
						return true
					}
				}
			}
		}
		return false
	}
	return true
}
