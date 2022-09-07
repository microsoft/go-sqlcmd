package sqlcmd

import (
	"os"
	"os/exec"
	"syscall"
)

func sysCommand(arg string) *exec.Cmd {
	cmd := exec.Command(comSpec())
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: cmd.Path + " " + comArgs(arg)}
	return cmd
}

// comSpec returns the path of the command shell executable
func comSpec() string {
	if cmd, ok := os.LookupEnv("COMSPEC"); ok {
		return cmd
	}
	return `C:\Windows\System32\cmd.exe`
}

func comArgs(args string) string {
	return `/c ` + args
}
