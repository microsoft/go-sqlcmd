package output

import (
	"github.com/microsoft/go-sqlcmd/internal/output/verbosity"
	"io"
)

type Options struct {
	OutputType     string
	LoggingLevel   verbosity.Level
	StandardWriter io.WriteCloser

	ErrorHandler func(err error)
	HintHandler  func(hints []string)

	unitTesting bool
}
