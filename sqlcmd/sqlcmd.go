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
	variables        *Variables
	//	lastBatch        []string
	//	batch            []string
	batchLine   int
	currentLine int
	//	db               *sql.DB
	out        io.WriteCloser
	noPassword bool
}

// New creates a new Sqlcmd instance
func New(l rline.IO, workingDirectory string, initialVariables *Variables, nopw bool) *Sqlcmd {
	return &Sqlcmd{
		lineIo:           l,
		workingDirectory: workingDirectory,
		noPassword:       nopw,
		variables:        initialVariables,
		currentLine:      1,
		batchLine:        1,
	}
}

func (s *Sqlcmd) Run() error {
	stdout, stderr, iactive := s.lineIo.Stdout(), s.lineIo.Stderr(), s.lineIo.Interactive()
	var lastError error
	for {
		if iactive {
			s.lineIo.Prompt(s.Prompt())
		}
		line, err := s.lineIo.Next()
		switch {
		case err == rline.ErrInterrupt:
			continue
		case err != nil:
			if err == io.EOF {
				return lastError
			}
			return err
		}
		isCommand, err := s.TryRunCommand(line)
		if err == ErrExitRequested {
			if s.out != nil {
				s.out.Close()
			}
			return nil
		}
		if err != nil {
			fmt.Fprintln(stderr, err)
			continue
		} else if isCommand {
			continue
		}
		fmt.Fprintln(stdout, string(line))
		break
	}
	return lastError
}

func (s *Sqlcmd) Prompt() string {
	return fmt.Sprint(s.batchLine) + ">"
}

func (s *Sqlcmd) TryRunCommand(line []rune) (bool, error) {
	return false, nil
}
