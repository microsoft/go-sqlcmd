package sqlcmd

import (
	"regexp"

	"github.com/microsoft/go-sqlcmd/errors"
)

type Command struct {
	// regex must include at least one group if it has parameters
	// Will be matched using FindStringSubmatch
	regex  *regexp.Regexp
	action func(*Sqlcmd, []string) error
	name   string
}

var Commands = []Command{
	{
		regex:  regexp.MustCompile(`(?im)^[\t ]*?:?QUIT(.*)$`),
		action: Quit,
		name:   "QUIT",
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

func Quit(s *Sqlcmd, args []string) error {
	if len(args) > 0 {
		return errors.InvalidCommandError("QUIT", 0)
	}
	return ErrExitRequested
}
