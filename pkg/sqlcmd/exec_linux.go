package sqlcmd

import (
	"os/exec"
)

func sysCommand(arg string) *exec.Cmd {
	cmd := exec.Command(comSpec(), "-c", arg)
	return cmd
}

// comSpec returns the path of the command shell executable
func comSpec() string {
	// /bin/sh will be a link to the shell
	return `/bin/sh`
}
