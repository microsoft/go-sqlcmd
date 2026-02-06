// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

//go:build !windows

package sqlcmd

import (
	"strconv"

	"github.com/microsoft/go-sqlcmd/internal/localizer"
	"golang.org/x/text/encoding"
)

// getSystemCodePageEncoding returns an error on non-Windows platforms
// since we don't have access to Windows API for codepage conversion.
// The built-in codepageRegistry covers the most common codepages.
// For additional codepages (e.g., Japanese EBCDIC), use Windows.
func getSystemCodePageEncoding(codepage int) (encoding.Encoding, error) {
	return nil, localizer.Errorf("unsupported codepage %s", strconv.Itoa(codepage))
}
