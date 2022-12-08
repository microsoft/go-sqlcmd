// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package verbosity

type Enum int

const (
	Error Enum = iota
	Warn
	Info
	Debug
	Trace
)
