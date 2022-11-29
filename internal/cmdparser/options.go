// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package cmdparser

type FlagOptions struct {
	Name      string
	Shorthand string
	Usage     string

	String        *string
	DefaultString string

	Int        *int
	DefaultInt int

	Bool        *bool
	DefaultBool bool
}

type Options struct {
	Aliases                    []string
	Examples                   []ExampleInfo
	FirstArgAlternativeForFlag *AlternativeForFlagInfo
	Long                       string
	Run                        func()
	Short                      string
	Use                        string
}
