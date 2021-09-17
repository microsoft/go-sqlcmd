// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package sqlcmd

const minCapIncrease = 512

// lineend is the slice to use when appending a line.
var lineend = []rune(SqlcmdEol)

// Batch provides the query text to run
type Batch struct {
	// read provides the next chunk of runes
	read batchScan
	// Buffer is the current batch text
	Buffer []rune
	// Length is the length of the statement
	Length int
	// raw is the unprocessed runes
	raw []rune
	// rawlen is the number of unprocessed runes
	rawlen int
	// quote indicates currently processing a quoted string
	quote rune
	// comment is the state of multi-line comment processing
	comment bool
	// batchline is the 1-based index of the next line.
	// Used for the prompt in interactive mode
	batchline int
	// linecount is the total number of batch lines processed in the session
	linecount uint
	// varmap tracks the location of expandable variables for the entire batch
	varmap map[int]string
	// linevarmap tracks the location of expandable variables on the current line
	linevarmap map[int]string
	// cmd is the set of Commands available
	cmd Commands
}

type batchScan func() (string, error)

// NewBatch creates a Batch which converts runes provided by reader into SQL batches
func NewBatch(reader batchScan, cmd Commands) *Batch {
	b := &Batch{
		read: reader,
		cmd:  cmd,
	}
	b.Reset(nil)
	return b
}

// String returns the current SQL batch text
func (b *Batch) String() string {
	return string(b.Buffer)
}

// Reset clears the current batch text and replaces it with new runes
func (b *Batch) Reset(r []rune) {
	b.Buffer, b.Length = nil, 0
	b.quote = 0
	b.comment = false
	b.batchline = 1
	if r != nil {
		b.raw, b.rawlen = r, len(r)
	} else {
		b.rawlen = 0
	}
	b.varmap = make(map[int]string)
}

// Next processes the next chunk of input and sets the Batch state accordingly.
// If the input contains a command to run, Next returns the Command and its
// parameters.
// Upon exit from Next, the caller can use the State method to determine if
// it represents a runnable SQL batch text.
func (b *Batch) Next() (*Command, []string, error) {
	b.linevarmap = nil
	var err error
	var i int
	if b.rawlen == 0 {
		s, err := b.read()
		if err != nil {
			return nil, nil, err
		}
		b.raw = []rune(s)
		b.rawlen = len(b.raw)
	}

	var command *Command
	var args []string
	var ok bool
	var scannedCommand bool
	b.linecount++
parse:
	for ; i < b.rawlen; i++ {
		c, next := b.raw[i], grab(b.raw, i+1, b.rawlen)
		switch {
		// we're in a quoted string
		case b.quote != 0:
			i, ok, err = b.readString(b.raw, i, b.rawlen, b.quote, b.linecount)
			if err != nil {
				break parse
			}
			if ok {
				b.quote = 0
			}
		// inside a multiline comment
		case b.comment:
			i, ok = readMultilineComment(b.raw, i, b.rawlen)
			b.comment = !ok
		// start of a string
		case c == '\'' || c == '"':
			b.quote = c
		// inline sql comment, skip to end of line
		case c == '-' && next == '-':
			i = b.rawlen
		// start a multi-line comment
		case c == '/' && next == '*':
			b.comment = true
			i++
		// continue processing quoted string or multiline comment
		case b.quote != 0 || b.comment:

		// Handle variable references
		case c == '$' && next == '(':
			vi, ok := readVariableReference(b.raw, i+2, b.rawlen)
			if ok {
				b.addVariableLocation(i, string(b.raw[i+2:vi]))
				i = vi

			} else {
				err = syntaxError(b.linecount)
				break parse
			}
		// Commands have to be alone on the line
		case !scannedCommand && b.cmd != nil:
			var cend int
			scannedCommand = true
			command, args, cend = readCommand(b.cmd, b.raw, i, b.rawlen)
			if command != nil {
				// remove the command from raw
				b.raw = append(b.raw[:i], b.raw[cend:]...)
				break parse
			}
		}
	}
	if err == nil {
		i = min(i, b.rawlen)
		empty := isEmptyLine(b.raw, 0, i)
		appendLine := b.quote != 0 || b.comment || !empty
		if !b.comment && command != nil && empty {
			appendLine = false
		}
		if appendLine {
			// any variables on the line need to be added to the global map
			inc := 0
			if b.Length > 0 {
				inc = len(lineend)
			}
			if b.linevarmap != nil {
				for v := range b.linevarmap {
					b.varmap[v+b.Length+inc] = b.linevarmap[v]
				}
			}
			// log.Printf(">> appending: `%s`", string(r[st:i]))
			b.append(b.raw[:i], lineend)
			b.batchline++
		}
		b.raw = b.raw[i:]
		b.rawlen = len(b.raw)
	} else {
		b.Reset(nil)
	}
	return command, args, err
}

// append appends r to b.Buffer separated by sep when b.Buffer is not already empty.
//
// Dynamically grows b.Buf as necessary to accommodate r and the separator.
// Specifically, when b.Buf is not empty, b.Buf will grow by increments of
// MinCapIncrease.
//
// After a call to append, b.Len will be len(b.Buf)+len(sep)+len(r). Call Reset
// to reset the Buf.
func (b *Batch) append(r, sep []rune) {
	rlen := len(r)
	// initial
	if b.Buffer == nil {
		b.Buffer, b.Length = r, rlen
		return
	}
	blen, seplen := b.Length, len(sep)
	tlen := blen + rlen + seplen
	// grow
	if bcap := cap(b.Buffer); tlen > bcap {
		n := tlen + 2*rlen
		n += minCapIncrease - (n % minCapIncrease)
		z := make([]rune, blen, n)
		copy(z, b.Buffer)
		b.Buffer = z
	}
	b.Buffer = b.Buffer[:tlen]
	copy(b.Buffer[blen:], sep)
	copy(b.Buffer[blen+seplen:], r)
	b.Length = tlen
}

// State returns a string representing the state of statement parsing.
// * Is in the middle of a multi-line comment
// - Has a non-empty batch ready to run
// = Is empty
// ' " Is in the middle of a multi-line quoted string
func (b *Batch) State() string {
	switch {
	case b.quote != 0:
		return string(b.quote)
	case b.comment:
		return "*"
	case b.Length != 0:
		return "-"
	}
	return "="
}

// readString seeks to the end of a string returning the position and whether
// or not the string's end was found.
//
// If the string's terminator was not found, then the result will be the passed
// end.
// An error is returned if the string contains a malformed variable reference
func (b *Batch) readString(r []rune, i, end int, quote rune, line uint) (int, bool, error) {
	var prev, c, next rune
	for ; i < end; i++ {
		c, next = r[i], grab(r, i+1, end)
		switch {
		case c == '$' && next == '(':
			vl, ok := readVariableReference(r, i+2, end)
			if ok {
				b.addVariableLocation(i, string(r[i+2:vl]))
				i = vl

			} else {
				return i, false, syntaxError(line)
			}
		case quote == '\'' && c == '\\':
			i++
			prev = 0
			continue
		case quote == '\'' && c == '\'' && next == '\'':
			i++
			continue
		case quote == '\'' && c == '\'' && prev != '\'',
			quote == '"' && c == '"':
			return i, true, nil
		}
		prev = c
	}
	return end, false, nil
}

// addVariableLocation is called for each variable on the current line
func (b *Batch) addVariableLocation(i int, v string) {
	if b.linevarmap == nil {
		b.linevarmap = make(map[int]string)
	}
	b.linevarmap[i] = v
}
