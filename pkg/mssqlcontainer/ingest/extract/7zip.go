package extract

import (
	"github.com/microsoft/go-sqlcmd/internal/container"
	"path/filepath"
	"regexp"
)

type sevenZip struct {
	controller  *container.Controller
	containerId string
}

func (e *sevenZip) Initialize(controller *container.Controller) {
	e.controller = controller
}

func (e *sevenZip) FileTypes() []string {
	return []string{"7z"}
}

func (e *sevenZip) IsInstalled(containerId string) bool {
	e.containerId = containerId

	return false
}

func (e *sevenZip) Extract(srcFile string, destFolder string) (string, string) {
	e.controller.RunCmdInContainer(e.containerId, []string{
		"/opt/7-zip/7zz",
		"x",
		"-aoa",
		"-o" + destFolder,
		"/var/opt/mssql/backup/" + srcFile,
	}, container.ExecOptions{})

	stdout, _, _ := e.controller.RunCmdInContainer(e.containerId, []string{
		"./opt/7-zip/7zz",
		"l",
		"-ba",
		"-slt",
		"/var/opt/mssql/backup/" + srcFile,
	}, container.ExecOptions{})

	var mdfFile string
	var ldfFile string

	paths := extractPaths(string(stdout))
	for _, p := range paths {
		if filepath.Ext(p) == ".mdf" {
			mdfFile = p
		}

		if filepath.Ext(p) == ".ldf" {
			ldfFile = p
		}
	}

	return mdfFile, ldfFile
}

func (e *sevenZip) Install() {
	e.controller.RunCmdInContainer(e.containerId, []string{
		"mkdir",
		"/opt/7-zip"}, container.ExecOptions{})

	e.controller.RunCmdInContainer(e.containerId, []string{
		"wget",
		"-O",
		"/opt/7-zip/7-zip.tar",
		"https://7-zip.org/a/7z2201-linux-x64.tar.xz"}, container.ExecOptions{})

	e.controller.RunCmdInContainer(e.containerId, []string{
		"tar",
		"xvf",
		"/opt/7-zip/7-zip.tar",
		"-C",
		"/opt/7-zip",
	}, container.ExecOptions{})

	e.controller.RunCmdInContainer(e.containerId, []string{
		"chmod",
		"u+x",
		"/opt/7-zip/7zz",
	}, container.ExecOptions{})
}

func extractPaths(input string) []string {
	re := regexp.MustCompile(`Path\s*=\s*(\S+)`)
	matches := re.FindAllStringSubmatch(input, -1)
	var paths []string
	for _, match := range matches {
		paths = append(paths, match[1])
	}
	return paths
}
