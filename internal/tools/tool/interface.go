// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package tool

type Tool interface {
	Init()
	Name() (name string)
	Run(args []string) (exitCode int, err error)
	IsInstalled() bool
	HowToInstall() string
}
