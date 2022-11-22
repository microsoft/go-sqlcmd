// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package formatter

import (
	"os"
	"testing"
)

func TestFormatter(t *testing.T) {

	s := "Hello"

	b := Base{
		StandardOutput: os.Stdout,
		ErrorHandlerCallback: func(err error) {
			if err != nil {
				panic(err)
			}
		},
	}

	j := Json{b}
	j.Serialize(s)

	x := Xml{b}
	x.Serialize(s)

	y := Yaml{b}
	y.Serialize(s)

}
