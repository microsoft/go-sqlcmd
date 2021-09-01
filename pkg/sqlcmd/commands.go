// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"syscall"
)

// Command defines a sqlcmd action which can be intermixed with the SQL batch
// Commands for sqlcmd are defined at https://docs.microsoft.com/sql/tools/sqlcmd-utility#sqlcmd-commands
type Command struct {
	// regex must include at least one group if it has parameters
	// Will be matched using FindStringSubmatch
	regex *regexp.Regexp
	// The function that implements the command. Third parameter is the line number
	action func(*Sqlcmd, []string, uint) error
	// Name of the command
	name string
}

// Commands is the set of Command implementations
var Commands = map[string]*Command{
	"QUIT": {
		regex:  regexp.MustCompile(`(?im)^[\t ]*?:?QUIT(?:[ \t]+(.*$)|$)`),
		action: quitCommand,
		name:   "QUIT",
	},
	"GO": {
		regex:  regexp.MustCompile(batchTerminatorRegex("GO")),
		action: goCommand,
		name:   "GO",
	},
	"OUT": {
		regex:  regexp.MustCompile(`(?im)^[ \t]*:OUT(?:[ \t]+(.*$)|$)`),
		action: outCommand,
		name:   "OUT",
	},
	"ERROR": {
		regex:  regexp.MustCompile(`(?im)^[ \t]*:ERROR(?:[ \t]+(.*$)|$)`),
		action: errorCommand,
		name:   "ERROR",
	},
}

func matchCommand(line string) (*Command, []string) {
	for _, cmd := range Commands {
		matchedCommand := cmd.regex.FindStringSubmatch(line)
		if matchedCommand != nil {
			return cmd, matchedCommand[1:]
		}
	}
	return nil, nil
}

func batchTerminatorRegex(terminator string) string {
	return fmt.Sprintf(`(?im)^[\t ]*?%s(?:[ ]+(.*$)|$)`, regexp.QuoteMeta(terminator))
}

// SetBatchTerminator attempts to set the batch terminator to the given value
// Returns an error if the new value is not usable in the regex
func SetBatchTerminator(terminator string) error {
	cmd := Commands["GO"]
	regex, err := regexp.Compile(batchTerminatorRegex(terminator))
	if err != nil {
		return err
	}
	cmd.regex = regex
	return nil
}

// quitCommand immediately exits the program without running any more batches
func quitCommand(s *Sqlcmd, args []string, line uint) error {
	if args != nil && strings.TrimSpace(args[0]) != "" {
		return InvalidCommandError("QUIT", line)
	}
	return ErrExitRequested
}

// goCommand runs the current batch the number of times specified
func goCommand(s *Sqlcmd, args []string, line uint) error {
	// default to 1 execution
	n := 1
	var err error
	if len(args) > 0 {
		cnt := strings.TrimSpace(args[0])
		if cnt != "" {
			_, err = fmt.Sscanf(cnt, "%d", &n)
		}
	}
	if err != nil || n < 1 {
		return InvalidCommandError("GO", line)
	}
	query := s.Query
	if query == "" {
		query = s.batch.String()
	}
	if query == "" {
		return nil
	}

	// This loop will likely be refactored to a helper when we implement -Q and :EXIT(query)
	for i := 0; i < n; i++ {

		s.Format.BeginBatch(query, s.vars, s.GetOutput(), s.GetError())
		rows, qe := s.db.Query(query)
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
	s.Query = ""
	s.batch.Reset(nil)
	return nil
}

// outCommand changes the output writer to use a file
func outCommand(s *Sqlcmd, args []string, line uint) error {
	switch {
	case strings.EqualFold(args[0], "stdout"):
		s.SetOutput(nil)
	case strings.EqualFold(args[0], "stderr"):
		s.SetOutput(os.NewFile(uintptr(syscall.Stderr), "/dev/stderr"))
	default:
		o, err := os.OpenFile(args[0], os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return InvalidFileError(err, args[0])
		}
		s.SetOutput(o)
	}
	return nil
}

// errorCommand changes the error writer to use a file
func errorCommand(s *Sqlcmd, args []string, line uint) error {
	switch {
	case strings.EqualFold(args[0], "stderr"):
		s.SetError(nil)
	case strings.EqualFold(args[0], "stdout"):
		s.SetError(os.NewFile(uintptr(syscall.Stderr), "/dev/stdout"))
	default:
		o, err := os.OpenFile(args[0], os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return InvalidFileError(err, args[0])
		}
		s.SetError(o)
	}
	return nil
}
