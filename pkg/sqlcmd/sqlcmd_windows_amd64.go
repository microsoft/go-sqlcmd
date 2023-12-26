package sqlcmd

import (
	"github.com/microsoft/go-mssqldb/msdsn"
	_ "github.com/microsoft/go-mssqldb/namedpipe"
	_ "github.com/microsoft/go-mssqldb/sharedmemory"
)

// Note: The order of includes above matters for namedpipe and sharedmemory.
// Go tools always sort by name.
// init() swaps shared memory protocol with np so it gets priority when dialing.

func init() {
	if len(msdsn.ProtocolParsers) == 4 {
		// reorder the protocol parsers to tcp->admin->lpc->np
		// Named pipes/shared memory package doesn't support ARM
		// Once there's a fix for https://github.com/natefinch/npipe/issues/34 incorporated into go-mssqldb,
		// reorder to lpc->admin->np->tcp
		var lpc = msdsn.ProtocolParsers[3]
		msdsn.ProtocolParsers[3] = msdsn.ProtocolParsers[2]
		msdsn.ProtocolParsers[2] = lpc
	}
}
