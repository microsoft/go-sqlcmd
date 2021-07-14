package sqlcmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"syscall"

	"github.com/microsoft/go-sqlcmd/sqlcmderrors"
)

type Command struct {
	// regex must include at least one group if it has parameters
	// Will be matched using FindStringSubmatch
	regex  *regexp.Regexp
	action func(*Sqlcmd, []string, uint) error
	name   string
}

var Commands = []Command{
	{
		regex:  regexp.MustCompile(`(?im)^[\t ]*?:?QUIT(?:[ \t]+(.*$)|$)`),
		action: Quit,
		name:   "QUIT",
	},
	{
		regex:  regexp.MustCompile(batchTerminatorRegex("GO")),
		action: Go,
		name:   "GO",
	},
	{
		regex:  regexp.MustCompile(`(?im)^[ \t]*:OUT(?:[ \t]+(.*$)|$)`),
		action: Out,
		name:   "OUT",
	},
	{
		regex:  regexp.MustCompile(`(?im)^[ \t]*:ERROR(?:[ \t]+(.*$)|$)`),
		action: Error,
		name:   "ERROR",
	},
}

func matchCommand(line string) (*Command, []string) {
	for _, cmd := range Commands {
		matchedCommand := cmd.regex.FindStringSubmatch(line)
		if matchedCommand != nil {
			return &cmd, matchedCommand[1:]
		}
	}
	return nil, nil
}

func batchTerminatorRegex(terminator string) string {
	return fmt.Sprintf(`(?im)^[\t ]*?%s(?:[ ]+(.*$)|$)`, regexp.QuoteMeta(terminator))
}

func SetBatchTerminator(terminator string) error {
	for i, cmd := range Commands {
		if cmd.name == "GO" {
			regex, err := regexp.Compile(batchTerminatorRegex(terminator))
			if err != nil {
				return err
			}
			Commands[i].regex = regex
			break
		}
	}
	return nil
}

func Quit(s *Sqlcmd, args []string, line uint) error {
	if args != nil && strings.TrimSpace(args[0]) != "" {
		return sqlcmderrors.InvalidCommandError("QUIT", line)
	}
	return ErrExitRequested
}

func Go(s *Sqlcmd, args []string, line uint) (err error) {
	// default to 1 execution
	var n int = 1
	if len(args) > 0 {
		cnt := strings.TrimSpace(args[0])
		if cnt != "" {
			_, err = fmt.Sscanf(cnt, "%d", &n)
		}
	}
	if err != nil || n < 1 {
		return sqlcmderrors.InvalidCommandError("GO", line)
	}
	for i := 0; i < n; i++ {
		fmt.Fprintf(s.GetOutput(), "GO: %s\n", s.batch.String())
	}
	s.batch.Reset(nil)
	return nil
}

func Out(s *Sqlcmd, args []string, line uint) (err error) {
	switch {
	case strings.EqualFold(args[0], "stdout"):
		s.SetOutput(nil)
	case strings.EqualFold(args[0], "stderr"):
		s.SetOutput(os.NewFile(uintptr(syscall.Stderr), "/dev/stderr"))
	default:
		o, err := os.OpenFile(args[0], os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return sqlcmderrors.InvalidFileError(err, args[0])
		}
		s.SetOutput(o)
	}
	return nil
}

func Error(s *Sqlcmd, args []string, line uint) error {
	switch {
	case strings.EqualFold(args[0], "stderr"):
		s.SetError(nil)
	case strings.EqualFold(args[0], "stdout"):
		s.SetError(os.NewFile(uintptr(syscall.Stderr), "/dev/stdout"))
	default:
		o, err := os.OpenFile(args[0], os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return sqlcmderrors.InvalidFileError(err, args[0])
		}
		s.SetError(o)
	}
	return nil
}
