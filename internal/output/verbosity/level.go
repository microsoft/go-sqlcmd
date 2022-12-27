// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package verbosity

// Level is an enumeration representing different verbosity levels for logging,
// ranging from Error to Trace. The values of the enumeration are
// Error, Warn, Info, Debug, and Trace, in increasing order of verbosity.
type Level int

const (
	Error Level = iota
	Warn
	Info
	Debug
	Trace
)
