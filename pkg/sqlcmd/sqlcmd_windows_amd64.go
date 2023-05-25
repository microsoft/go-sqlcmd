package sqlcmd

import (
	"github.com/microsoft/go-mssqldb/msdsn"
	_ "github.com/microsoft/go-mssqldb/namedpipe"
	_ "github.com/microsoft/go-mssqldb/sharedmemory"
)

func init() {
	if len(msdsn.ProtocolParsers) == 4 {
		// reorder the protocol parsers to lpc->admin->tcp->np
		// ODBC follows this same order.
		var tcp = msdsn.ProtocolParsers[0]
		msdsn.ProtocolParsers[0] = msdsn.ProtocolParsers[3]
		msdsn.ProtocolParsers[3] = msdsn.ProtocolParsers[2]
		msdsn.ProtocolParsers[2] = tcp
	}
}
