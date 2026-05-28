// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestHelperProcess is not a real test; it is exec'd by the stubbed
// execCommand to emit canned vswhere output. It echoes its first argument to
// stdout and exits non-zero when a second argument "fail" is present.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	for i, a := range args {
		if a == "--" {
			args = args[i+1:]
			break
		}
	}
	if len(args) >= 2 && args[1] == "fail" {
		os.Exit(1)
	}
	if len(args) >= 1 {
		os.Stdout.WriteString(args[0])
	}
	os.Exit(0)
}

// stubVswhere makes vswhereFind's preconditions pass (ProgramFiles(x86) set and
// vswhere.exe present) and replaces execCommand with a helper-process stub that
// returns output / fails as requested while capturing the args vswhere is
// called with.
func stubVswhere(t *testing.T, output string, fail bool) *[]string {
	t.Helper()

	pf86 := t.TempDir()
	installer := filepath.Join(pf86, "Microsoft Visual Studio", "Installer")
	if err := os.MkdirAll(installer, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(installer, "vswhere.exe"), []byte("stub"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ProgramFiles(x86)", pf86)

	var captured []string
	orig := execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		captured = args
		helper := []string{"-test.run=TestHelperProcess", "--", output}
		if fail {
			helper = append(helper, "fail")
		}
		cmd := exec.Command(os.Args[0], helper...)
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		return cmd
	}
	t.Cleanup(func() { execCommand = orig })

	return &captured
}

func TestVswhereFindLatestWhenVersionEmpty(t *testing.T) {
	args := stubVswhere(t, "C:\\VS\\SSMS", false)

	got := vswhereFind("Microsoft.VisualStudio.Product.Ssms", "")

	if got != "C:\\VS\\SSMS" {
		t.Errorf("expected install path, got %q", got)
	}
	if !contains(*args, "-latest") {
		t.Errorf("expected -latest in args, got %v", *args)
	}
	if contains(*args, "-version") {
		t.Errorf("did not expect -version when version is empty, got %v", *args)
	}
}

func TestVswhereFindRestrictsToMajorVersionLine(t *testing.T) {
	args := stubVswhere(t, "C:\\VS\\SSMS21", false)

	got := vswhereFind("Microsoft.VisualStudio.Product.Ssms", "21")

	if got != "C:\\VS\\SSMS21" {
		t.Errorf("expected install path, got %q", got)
	}
	if !contains(*args, "[21.0,22.0)") {
		t.Errorf("expected version range [21.0,22.0) in args, got %v", *args)
	}
	if contains(*args, "-latest") {
		t.Errorf("did not expect -latest when version is pinned, got %v", *args)
	}
}

func TestVswhereFindReturnsFirstNonEmptyLine(t *testing.T) {
	stubVswhere(t, "\n  C:\\VS\\First  \nC:\\VS\\Second\n", false)

	if got := vswhereFind("Microsoft.VisualStudio.Product.Ssms", ""); got != "C:\\VS\\First" {
		t.Errorf("expected first non-empty trimmed line, got %q", got)
	}
}

func TestVswhereFindReturnsEmptyOnError(t *testing.T) {
	stubVswhere(t, "C:\\VS\\SSMS", true)

	if got := vswhereFind("Microsoft.VisualStudio.Product.Ssms", ""); got != "" {
		t.Errorf("expected empty string when vswhere fails, got %q", got)
	}
}

func TestVswhereFindReturnsEmptyWhenInstallerMissing(t *testing.T) {
	t.Setenv("ProgramFiles(x86)", t.TempDir()) // no vswhere.exe inside

	if got := vswhereFind("Microsoft.VisualStudio.Product.Ssms", ""); got != "" {
		t.Errorf("expected empty string when vswhere.exe is absent, got %q", got)
	}
}

func contains(s []string, want string) bool {
	for _, v := range s {
		if strings.Contains(v, want) {
			return true
		}
	}
	return false
}
