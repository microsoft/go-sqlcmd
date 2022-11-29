// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package formatter

import "io"

type Base struct {
	StandardOutput       io.WriteCloser
	ErrorHandlerCallback func(err error)
}

func (f *Base) CheckErr(err error) {
	if f.ErrorHandlerCallback == nil {
		panic("errorHandlerCallback not initialized")
	}

	f.ErrorHandlerCallback(err)
}

func (f *Base) Output(bytes []byte) {
	_, err := f.StandardOutput.Write(bytes)
	f.CheckErr(err)
}
