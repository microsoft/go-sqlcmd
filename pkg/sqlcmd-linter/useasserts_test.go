package sqlcmdlinter

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestUseAsserts(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Errorf("Failed to get wd: %s", err)
	}
	analysistest.Run(t, filepath.Join(wd, `testdata`), AssertAnalyzer, "useassert_linter_tests/...")
}
