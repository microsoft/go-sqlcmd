// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package console

import (
	"bufio"
	"os"

	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
	"github.com/peterh/liner"
)

type console struct {
	impl            *liner.State
	historyFile     string
	prompt          string
	stdinRedirected bool
	stdinReader     *bufio.Reader
}

// NewConsole creates a sqlcmdConsole implementation that provides these features:
// - Storage of input history to a local file. History can be scrolled through using the up and down arrow keys.
// - Simple tab key completion of SQL keywords
func NewConsole(historyFile string) sqlcmd.Console {
	c := &console{
		impl:            liner.NewLiner(),
		historyFile:     historyFile,
		stdinRedirected: isStdinRedirected(),
	}

	if c.stdinRedirected {
		c.stdinReader = bufio.NewReader(os.Stdin)
	} else {
		c.impl.SetCtrlCAborts(true)
		c.impl.SetCompleter(CompleteLine)
		if c.historyFile != "" {
			if f, err := os.Open(historyFile); err == nil {
				_, _ = c.impl.ReadHistory(f)
				f.Close()
			}
		}
	}
	return c
}

// Close writes out the history data to disk and closes the console buffers
func (c *console) Close() {
	if !c.stdinRedirected && c.historyFile != "" {
		if f, err := os.Create(c.historyFile); err == nil {
			_, _ = c.impl.WriteHistory(f)
			f.Close()
		}
	}

	if !c.stdinRedirected {
		c.impl.Close()
	}
}

// Readline displays the current prompt and returns a line of text entered by the user.
// It appends the returned line to the history buffer.
// If the user presses Ctrl-C the error returned is sqlcmd.ErrCtrlC
// If stdin is redirected, it reads directly from stdin without displaying prompts
func (c *console) Readline() (string, error) {
	// Handle redirected stdin without displaying prompts
	if c.stdinRedirected {
		line, err := c.stdinReader.ReadString('\n')
		if err != nil {
			return "", err
		}
		// Trim the trailing newline
		if len(line) > 0 && line[len(line)-1] == '\n' {
			line = line[:len(line)-1]
			// Also trim carriage return if present
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
		}
		return line, nil
	}

	// Interactive terminal mode with prompts
	s, err := c.impl.Prompt(c.prompt)
	if err == liner.ErrPromptAborted {
		return "", sqlcmd.ErrCtrlC
	}
	c.impl.AppendHistory(s)
	return s, err
}

// ReadPassword displays the given prompt and returns the password entered by the user.
// If the user presses Ctrl-C the error returned is sqlcmd.ErrCtrlC
func (c *console) ReadPassword(prompt string) ([]byte, error) {
	// Even when stdin is redirected, we need to use the prompt for passwords
	// since they should not be read from the redirected input
	b, err := c.impl.PasswordPrompt(prompt)
	if err == liner.ErrPromptAborted {
		return []byte{}, sqlcmd.ErrCtrlC
	}
	return []byte(b), err
}

// SetPrompt sets the prompt text shown to input the next line
func (c *console) SetPrompt(s string) {
	c.prompt = s
}
