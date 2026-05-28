// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestSSMSSearchLocationsUsesVswhereRoot(t *testing.T) {
	root := t.TempDir()
	exe := filepath.Join(root, "Common7", "IDE", "Ssms.exe")
	if err := os.MkdirAll(filepath.Dir(exe), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exe, []byte("stub"), 0600); err != nil {
		t.Fatal(err)
	}

	stubVswhereReturning(t, root)

	ssms := SSMS{}
	ssms.Init()
	ssms.SetVersion("21")

	if !ssms.IsInstalled() {
		t.Errorf("expected SSMS to be reported installed at %q", exe)
	}
}

func TestSSMSNotInstalledWhenVswhereEmpty(t *testing.T) {
	stubVswhereReturning(t, "")

	ssms := SSMS{}
	ssms.Init()
	ssms.SetVersion("21")

	if ssms.IsInstalled() {
		t.Error("expected SSMS to be reported not installed when vswhere finds nothing")
	}
}

// stubVswhereReturning makes vswhereFind succeed and emit the given install
// root (empty root simulates "no instance found").
func stubVswhereReturning(t *testing.T, root string) {
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

	orig := execCommand
	execCommand = func(command string, args ...string) *exec.Cmd {
		helper := []string{"-test.run=TestHelperProcess", "--", root}
		cmd := exec.Command(os.Args[0], helper...)
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		return cmd
	}
	t.Cleanup(func() { execCommand = orig })
}
