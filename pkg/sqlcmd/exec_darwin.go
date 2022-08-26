package sqlcmd

import (
	"os/exec"
)

func sysCommand(arg string) *exec.Cmd {
	cmd := exec.Command(comSpec(), comArgs(arg))
	return cmd
}

// comSpec returns the path of the command shell executable
func comSpec() string {
	// /bin/sh will be a link to the shell
	return `/bin/sh`
}

func comArgs(args string) string {
	return args
}
