package sqlcmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/microsoft/go-sqlcmd/errors"
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
	for _, cmd := range Commands {
		if cmd.name == "GO" {
			regex, err := regexp.Compile(batchTerminatorRegex(terminator))
			if err != nil {
				return err
			}
			cmd.regex = regex
			break
		}
	}
	return nil
}

func Quit(s *Sqlcmd, args []string, line uint) error {
	if args != nil && strings.TrimSpace(args[0]) != "" {
		return errors.InvalidCommandError("QUIT", line)
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
		return errors.InvalidCommandError("GO", line)
	}
	for i := 0; i < n; i++ {
		fmt.Fprintf(s.lineIo.Stdout(), "GO: %s\n", s.batch.String())
	}
	s.batch.Reset(nil)
	return nil
}
