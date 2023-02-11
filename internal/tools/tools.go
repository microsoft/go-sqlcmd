package tools

import (
	"github.com/microsoft/go-sqlcmd/internal/tools/tool"
)

var tools = []tool.Tool{
	&tool.AzureDataStudio{},
}
