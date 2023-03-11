// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package main

import (
	_ "github.com/alecthomas/chroma" // want "Non-internal packages should not import \"github.com/alecthomas/chroma\""
	_ "github.com/microsoft/go-sqlcmd/pkg/sqlcmd"
	_ "github.com/stretchr/testify/assert"
)

var X = 1
