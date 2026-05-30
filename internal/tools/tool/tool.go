// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/microsoft/go-sqlcmd/internal/io/file"
)

// earlyExitWindow is how long Run waits after Start before considering the
// launched tool successfully detached. If the child exits within this window
// Run returns its exit code and error so callers can surface install help or
// other troubleshooting hints instead of silently reporting success.
const earlyExitWindow = 750 * time.Millisecond

func (t *tool) Init() {
	panic("Do not call directly")
}

func (t *tool) Name() string {
	return t.description.Name
}

func (t *tool) SetExePathAndName(exeName string) {
	t.exeName = exeName
}

// ExePath returns the resolved executable path, or "" if the tool was not
// found. Valid only after Init (and SetBuild, where supported).
func (t *tool) ExePath() string {
	return t.exeName
}

func (t *tool) SetToolDescription(description Description) {
	t.description = description
}

func (t *tool) IsInstalled() bool {
	if t.installed != nil {
		return *t.installed
	}

	t.installed = new(bool)
	// Handle case where tool wasn't found during Init (exeName is empty)
	if t.exeName != "" && file.Exists(t.exeName) {
		*t.installed = true
	} else {
		*t.installed = false
	}

	return *t.installed
}

func (t *tool) HowToInstall() string {
	var sb strings.Builder

	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("%q is not installed on this machine.\n\n", t.description.Name))
	sb.WriteString(fmt.Sprintf("%v\n\n", t.description.Purpose))
	sb.WriteString(fmt.Sprintf("To install %q...\n\n%v\n", t.description.Name, t.description.InstallText))

	return sb.String()
}

func (t *tool) Run(args []string) (int, error) {
	if t.installed == nil {
		return 1, fmt.Errorf("internal error: Call IsInstalled before Run")
	}

	cmd := t.generateCommandLine(args)

	// Normalize argv[0] so it matches cmd.Path on every platform. The darwin
	// implementation builds Args starting at "-a" because it shells out via
	// /usr/bin/open; without this fix the launched process sees "-a" as argv[0]
	// and mis-parses subsequent flags.
	if len(cmd.Args) == 0 || cmd.Args[0] != cmd.Path {
		cmd.Args = append([]string{cmd.Path}, cmd.Args...)
	}

	// Redirect stdio to the null device so exec.Cmd does not spawn pipe
	// drainer goroutines. Without this, Start leaves goroutines blocked on
	// the child's stdout/stderr until the GUI tool exits, which keeps
	// sqlcmd's process tree alive even after Process.Release. If opening
	// the null device fails, fall back to inheriting the parent's stdio
	// (also goroutine-free) rather than leaving the bytes.Buffer pipes
	// generateCommandLine attached.
	if devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0); err == nil {
		cmd.Stdin = devNull
		cmd.Stdout = devNull
		cmd.Stderr = devNull
		defer func() { _ = devNull.Close() }()
	} else {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Start(); err != nil {
		return 1, err
	}

	// Wait briefly for the child to fail fast (invalid args, missing
	// dependency). If it survives the window, treat the launch as successful
	// and let the GUI tool keep running; the Wait goroutine dies with sqlcmd
	// and the child is reparented (Unix) or unaffected (Windows).
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), exitErr
		}
		if err != nil {
			return 1, err
		}
		return 0, nil
	case <-time.After(earlyExitWindow):
		return 0, nil
	}
}
