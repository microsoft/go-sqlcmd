// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmdlinter

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var AssertAnalyzer = &analysis.Analyzer{
	Name:     "assertlint",
	Doc:      "Require use of asserts instead of fmt.Error functions in tests",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runAsserts,
}

var blockedTestingMethods = []string{"Error", "ErrorF", "Fail", "FailNow", "Fatal", "Fatalf"}

func runAsserts(pass *analysis.Pass) (interface{}, error) {
	// pass.ResultOf[inspect.Analyzer] will be set if we've added inspect.Analyzer to Requires.
	// Analyze code and make an AST from the file:
	inspectorInstance := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}

	inspectorInstance.Preorder(nodeFilter, func(n ast.Node) {
		node := n.(*ast.CallExpr)
		switch fun := node.Fun.(type) {
		case (*ast.SelectorExpr):
			switch funX := fun.X.(type) {
			case (*ast.Ident):
				if funX.Name == "t" && contains(blockedTestingMethods, fun.Sel.Name) {
					pass.Reportf(node.Pos(), "Use assert package methods instead of %s", fun.Sel.Name)
				}
			default:
				return
			}
		case (*ast.Ident):
			if fun.Name == "recover" {
				pass.Reportf(node.Pos(), "Use assert.Panics instead of recover()")
			}
		}
	})
	return nil, nil
}

func contains(a []string, v string) bool {
	for _, val := range a {
		if val == v {
			return true
		}
	}
	return false
}
