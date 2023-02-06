package tool

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"runtime"
)

type Base struct {
	name              string
	isInstalledCalled bool
	installed         bool
	lookPathError     error
	exeName           string
	exeFullPath       string
	toolYaml          ToolDescription
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

func (t *Base) SetToolYaml(toolYaml ToolDescription) {
	t.toolYaml = toolYaml
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

func (t *Base) HowToInstall() {
	var text string
	switch runtime.GOOS {
	case "windows":
		text = t.toolYaml.InstallText.Windows
	case "darwin":
		text = t.toolYaml.InstallText.Mac
	case "linux":
		text = t.toolYaml.InstallText.Linux
	default:
		panic(fmt.Sprintf("Not a supported platform (%v)", runtime.GOOS))
	}

	fmt.Printf("\n\n")
	fmt.Printf("WARNING: '%v' is not installed on this machine.\n\n", t.name)
	fmt.Printf("%v\n\n", t.toolYaml.Purpose)
	fmt.Printf("To install '%v'...\n\n%v\n", t.name, text)
}

func (t *Base) Run(args []string) (int, error, string, string) {
	if !t.isInstalledCalled {
		panic("Call IsInstalled before Run")
	}

	if !t.installed {
		log.Fatal(fmt.Sprintf("The '%v' tool is not found on this machine. "+
			"Please install it. (%v)", t.Name(), t.lookPathError))
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

	return cmd.ProcessState.ExitCode(), err, stdout.String(), stderr.String()
}
