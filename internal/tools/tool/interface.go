// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

type Tool interface {
	Init()
	Name() (name string)
	Run(args []string, options RunOptions) (exitCode int, err error)
	IsInstalled() bool
	HowToInstall() string
}

type RunOptions struct {
	Interactive bool
}
