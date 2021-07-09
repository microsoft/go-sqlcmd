package sqlcmd

const MinCapIncrease = 512

// lineend is the slice to use when appending a line.
var lineend = []rune{'\n'}

type Batch struct {
	// read provides the next chunk of runes
	read func() ([]rune, error)
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
}

func NewBatch(reader func() ([]rune, error)) *Batch {
	b := &Batch{
		read: reader,
	}
	return b
}

func (b *Batch) String() string {
	return string(b.Buffer)
}

func (b *Batch) Reset(r []rune) {
	b.Buffer, b.Length = nil, 0
	b.quote = 0
	b.comment = false
	b.batchline = 1
	if r != nil {
		b.raw, b.rawlen = r, len(r)
	}
}

func (b *Batch) Next() (*Command, []string, error) {
	var err error
	var i int
	if b.rawlen == 0 {
		b.raw, err = b.read()
		if err != nil {
			return nil, nil, err
		}
		b.rawlen = len(b.raw)
	}

	var command *Command
	var args []string
	var ok bool
	var scannedCommand bool
parse:
	for ; i < b.rawlen; i++ {
		c, next := b.raw[i], grab(b.raw, i+1, b.rawlen)
		switch {
		// we're in a quoted string
		case b.quote != 0:
			i, ok = readString(b.raw, i, b.rawlen, b.quote)
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
		// Commands have to be alone on the line
		case !scannedCommand:
			var cend int
			scannedCommand = true
			command, args, cend = readCommand(b.raw, i, b.rawlen)
			if command != nil {
				// remove the command from raw
				b.raw = append(b.raw[:i], b.raw[cend:]...)
				break parse
			}
		}
	}
	i = min(i, b.rawlen)
	empty := isEmptyLine(b.raw, 0, i)
	appendLine := b.quote != 0 || b.comment || !empty
	if !b.comment && command != nil && empty {
		appendLine = false
	}
	if appendLine {
		// skip leading space when empty
		st := 0
		if b.Length == 0 {
			st, _ = findNonSpace(b.raw, 0, i)
		}
		// log.Printf(">> appending: `%s`", string(r[st:i]))
		b.Append(b.raw[st:i], lineend)
	}
	b.raw = b.raw[i:]
	b.rawlen = len(b.raw)
	return command, args, nil
}

// Append appends r to b.Buffer separated by sep when b.Buffer is not already empty.
//
// Dynamically grows b.Buf as necessary to accommodate r and the separator.
// Specifically, when b.Buf is not empty, b.Buf will grow by increments of
// MinCapIncrease.
//
// After a call to Append, b.Len will be len(b.Buf)+len(sep)+len(r). Call Reset
// to reset the Buf.
func (b *Batch) Append(r, sep []rune) {
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
		n += MinCapIncrease - (n % MinCapIncrease)
		z := make([]rune, blen, n)
		copy(z, b.Buffer)
		b.Buffer = z
	}
	b.Buffer = b.Buffer[:tlen]
	copy(b.Buffer[blen:], sep)
	copy(b.Buffer[blen+seplen:], r)
	b.Length = tlen
	b.linecount++
}

// State returns a string representing the state of statement parsing.
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
