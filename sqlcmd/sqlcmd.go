package sqlcmd

import (
	"errors"
	"fmt"
	"io"

	"github.com/xo/usql/rline"
)

var (
	ErrExitRequested = errors.New("exit")
)

// Sqlcmd is the core processor for text lines.
//
// It accumulates non-command lines in a buffer and  and sends command lines to the appropriate command runner.
// When the batch delimiter is encountered it sends the current batch to the active connection and prints
// the results to the output writer

type Sqlcmd struct {
	lineIo           rline.IO
	workingDirectory string
	//	db               *sql.DB
	out   io.WriteCloser
	batch *Batch
	// Exitcode is returned to the operating system when the process exits
	Exitcode int
}

// New creates a new Sqlcmd instance
func New(l rline.IO, workingDirectory string) *Sqlcmd {
	return &Sqlcmd{
		lineIo:           l,
		workingDirectory: workingDirectory,
		batch:            &Batch{read: l.Next},
	}
}

func (s *Sqlcmd) Run() error {
	stdout, stderr, iactive := s.lineIo.Stdout(), s.lineIo.Stderr(), s.lineIo.Interactive()
	var lastError error
	for {
		var execute bool
		if iactive {
			s.lineIo.Prompt(s.Prompt())
		}
		cmd, args, err := s.batch.Next()
		switch {
		case err == rline.ErrInterrupt:
			s.batch.Reset(nil)
			continue
		case err != nil:
			if err == io.EOF {
				if s.batch.Length == 0 {
					return lastError
				} else {
					execute = true
				}
			} else {
				return err
			}
		}
		if cmd != nil {
			err = s.RunCommand(cmd, args)
			if err == ErrExitRequested {
				if s.out != nil {
					s.out.Close()
				}
				break
			}
			if err != nil {
				fmt.Fprintln(stderr, err)
				lastError = err
				continue
			}
		}
		if execute {
			fmt.Fprintln(stdout, "Execute: "+s.batch.String())
			s.batch.Reset(nil)
		}
	}
	return lastError
}

func (s *Sqlcmd) Prompt() string {
	ch := ">"
	if s.batch.quote != 0 || s.batch.comment {
		ch = "~"
	}
	return fmt.Sprint(s.batch.batchline) + ch
}

func (s *Sqlcmd) RunCommand(cmd *Command, args []string) error {
	return cmd.action(s, args, s.batch.linecount)
}
