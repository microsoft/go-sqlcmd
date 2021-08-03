// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

// Commands for sqlcmd are defined at https://docs.microsoft.com/sql/tools/sqlcmd-utility#sqlcmd-commands
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
	regex *regexp.Regexp
	// The function that implements the command. Third parameter is the line number
	action func(*Sqlcmd, []string, uint) error
	// Name of the command
	name string
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

// Attempts to set the batch terminator to the given value
// Returns an error if the new value is not usable in the regex
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

// Immediately exits the program without running any more batches
func Quit(s *Sqlcmd, args []string, line uint) error {
	if args != nil && strings.TrimSpace(args[0]) != "" {
		return sqlcmderrors.InvalidCommandError("QUIT", line)
	}
	return ErrExitRequested
}

// Runs the current batch the number of times specified
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

	// This loop will likely be refactored to a helper when we implement -Q and :EXIT(query)
	for i := 0; i < n; i++ {
		s.Format.BeginBatch(s.batch.String(), s.vars, s.GetOutput(), s.GetError())
		rows, qe := s.db.Query(s.batch.String())
		if qe != nil {
			s.Format.AddError(qe)
		}

		results := true
		for qe == nil && results {
			cols, err := rows.ColumnTypes()
			if err != nil {
				s.Format.AddError(err)
			} else {
				s.Format.BeginResultSet(cols)
				active := rows.Next()
				for active {
					s.Format.AddRow(rows)
					active = rows.Next()
				}
				if err = rows.Err(); err != nil {
					s.Format.AddError(err)
				}
				s.Format.EndResultSet()
			}
			results = rows.NextResultSet()
			if err = rows.Err(); err != nil {
				s.Format.AddError(err)
			}
		}
		s.Format.EndBatch()
	}
	s.batch.Reset(nil)
	return nil
}

// Changes the output writer to use a file
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

// Changes the error writer to use a file
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
