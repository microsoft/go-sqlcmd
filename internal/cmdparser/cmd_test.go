package cmdparser

import (
	"github.com/spf13/cobra"
	"testing"
)

func TestCmd_run(t *testing.T) {
	s := ""
	c := Cmd{
		options: Options{
			FirstArgAlternativeForFlag: &AlternativeForFlagInfo{
				Flag:  "name",
				Value: &s,
			},
		},
		command: cobra.Command{},
	}
	c.AddFlag(FlagOptions{
		Name:   "name",
		Usage:  "name",
		String: &s,
	})
	c.run(nil, []string{"name-value"})
}
