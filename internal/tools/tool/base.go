package tool

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

type Base struct {
	name              string
	isInstalledCalled bool
	installed         bool
	lookPathError     error
	exeName           string
	exeFullPath       string
	description       Description
}

func (t *Base) Init() {
	panic("Do not call directly")
}

func (t *Base) Name() string {
	return t.name
}

func (t *Base) SetName(name string) {
	t.name = name
}

func (t *Base) ExeName() string {
	return t.exeName
}

func (t *Base) ExeFullPath() string {
	return t.exeName
}

func (t *Base) SetExeName(exeName string) {
	t.exeName = exeName
}

func (t *Base) SetToolDescription(description Description) {
	t.description = description
}

func (t *Base) Where() string {
	return t.exeFullPath
}

func (t *Base) IsInstalled() bool {
	if t.isInstalledCalled {
		return t.installed
	}

	if t.exeName == "" {
		panic("exeName is empty")
	}

	t.exeFullPath, t.lookPathError = exec.LookPath(t.exeName)

	if t.lookPathError == nil {
		t.installed = true
	}

	t.isInstalledCalled = true

	return t.installed
}

func (t *Base) HowToInstall() string {
	var text string
	switch runtime.GOOS {
	case "windows":
		text = t.description.InstallText.Windows
	case "darwin":
		text = t.description.InstallText.Mac
	case "linux":
		text = t.description.InstallText.Linux
	default:
		panic(fmt.Sprintf("Not a supported platform (%v)", runtime.GOOS))
	}

	var sb strings.Builder

	sb.WriteString("\n\n")
	sb.WriteString(fmt.Sprintf("WARNING: %q is not installed on this machine.\n\n", t.name))
	sb.WriteString(fmt.Sprintf("%v\n\n", t.description.Purpose))
	sb.WriteString(fmt.Sprintf("To install '%v'...\n\n%v\n", t.name, text))

	return sb.String()
}

func (t *Base) Run(args []string) (int, error) {
	if !t.isInstalledCalled {
		panic("Call IsInstalled before Run")
	}

	// args requires the .exeFullPath to be arg[0], so prepend it
	args = append([]string{t.exeFullPath}, args...)

	var stdout, stderr bytes.Buffer
	cmd := &exec.Cmd{
		Path:   t.exeFullPath,
		Args:   args,
		Stdout: &stdout,
		Stderr: &stderr,
	}

	err := cmd.Run()

	return cmd.ProcessState.ExitCode(), err
}
