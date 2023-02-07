package tools

import (
	"fmt"
	"github.com/microsoft/go-sqlcmd/internal/tools/tool"
)

var calledBefore bool

func NewTool(toolName string) (tool tool.Tool) {
	if !calledBefore {
		// Init all the tools
		for _, t := range tools {
			t.Init()
		}
		calledBefore = true
	}

	// Return the asked for tool
	for _, t := range tools {
		if t.Name() == toolName {
			tool = t
		}
	}

	if tool == nil {
		panic(fmt.Sprintf("Tool %q is not a supported tool", toolName))
	}

	return tool
}
