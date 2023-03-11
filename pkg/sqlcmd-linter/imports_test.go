// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmdlinter

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestImports(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("Failed to get wd: %s", err)
	}
	analysistest.Run(t, filepath.Join(wd, `testdata`), ImportsAnalyzer, "imports_linter_tests/...")
}
