// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package output

import (
	"github.com/microsoft/go-sqlcmd/internal/output/formatter"
	"github.com/microsoft/go-sqlcmd/internal/output/verbosity"
	"io"
)

type Output struct {
	errorCallback func(err error)
	hintCallback  func(hints []string)

	formatter           formatter.Formatter
	loggingLevel        verbosity.Level
	standardWriteCloser io.WriteCloser
}
