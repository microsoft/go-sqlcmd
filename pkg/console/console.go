// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package console

import (
	"os"

	"github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
	"github.com/peterh/liner"
)

type console struct {
	impl        *liner.State
	historyFile string
	prompt      string
}

// NewConsole creates a sqlcmdConsole implementation that provides these features:
// - Storage of input history to a local file. History can be scrolled through using the up and down arrow keys.
// - Simple tab key completion of SQL keywords
func NewConsole(historyFile string) sqlcmd.Console {
	c := &console{
		impl:        liner.NewLiner(),
		historyFile: historyFile,
	}
	c.impl.SetCtrlCAborts(true)
	if c.historyFile != "" {
		if f, err := os.Open(historyFile); err == nil {
			_, _ = c.impl.ReadHistory(f)
			f.Close()
		}
	}
	return c
}

// Close writes out the history data to disk and closes the console buffers
func (c *console) Close() {
	if c.historyFile != "" {
		if f, err := os.Create(c.historyFile); err == nil {
			_, _ = c.impl.WriteHistory(f)
			f.Close()
		}
	}
	c.impl.Close()
}

// Readline displays the current prompt and returns a line of text entered by the user.
// It appends the returned line to the history buffer.
// If the user presses Ctrl-C the error returned is sqlcmd.ErrCtrlC
func (c *console) Readline() (string, error) {
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

/*
IDS_AUX_KEYWORDS,                           "BREAK,BROWSE,BULK,CHECKPOINT,CLUSTERED,COMMITTED,COMPUTE,CONFIRM,CONTROLROW,DATABASE,DBCC,DISK,DISTRIBUTED,DUMMY,DUMP,ERRLVL,ERROREXIT,EXIT,FILE,FILLFACTOR,FLOPPY,HOLDLOCK,IDENTITY_INSERT,IDENTITYCOL,IF,KILL,LINENO,LOAD,MIRROREXIT,"
IDS_AUX_KEYWORDS+1,                         "NONCLUSTERED,OFF,OFFSETS,ONCE,OVER,PERCENT,PERM,PERMANENT,PLAN,PRINT,PROC,PROCESSEXIT,RAISERROR,READ,READTEXT,RECONFIGURE,REPEATABLE,RETURN,ROWCOUNT,RULE,SAVE,SERIALIZABLE,SETUSER,SHUTDOWN,STATISTICS,"
IDS_AUX_KEYWORDS+2,                         "TAPE,TEMP,TEXTSIZE,TOP,TRAN,TRIGGER,TRUNCATE,TSEQUEL,UNCOMMITTED,UPDATETEXT,USE,WAITFOR,WHILE,WRITETEXT"
IDS_AUX_KEYWORDS+3,                         ""  //Empty string to terminate the keyword list

// katmai key words. DUMP and LOAD are removed. BACKUP and RESTORE are added.
IDS_AUX_KEYWORDS_KATMAI,                    "BACKUP,BREAK,BROWSE,BULK,CHECKPOINT,CLUSTERED,COMMITTED,COMPUTE,CONFIRM,CONTROLROW,DATABASE,DBCC,DISK,DISTRIBUTED,DUMMY,ERRLVL,ERROREXIT,EXIT,FILE,FILLFACTOR,FLOPPY,HOLDLOCK,IDENTITY_INSERT,IDENTITYCOL,IF,KILL,LINENO,MERGE,MIRROREXIT,"
IDS_AUX_KEYWORDS_KATMAI+1,                  "NONCLUSTERED,OFF,OFFSETS,ONCE,OVER,PERCENT,PERM,PERMANENT,PLAN,PRINT,PROC,PROCESSEXIT,RAISERROR,READ,READTEXT,RECONFIGURE,REPEATABLE,RESTORE,RETURN,ROWCOUNT,RULE,SAVE,SERIALIZABLE,SETUSER,SHUTDOWN,STATISTICS,"
IDS_AUX_KEYWORDS_KATMAI+2,                  "TAPE,TEMP,TEXTSIZE,TOP,TRAN,TRIGGER,TRUNCATE,TSEQUEL,UNCOMMITTED,UPDATETEXT,USE,WAITFOR,WHILE,WRITETEXT"
IDS_AUX_KEYWORDS_KATMAI+3,                  ""  //Empty string to terminate the keyword list
*/
